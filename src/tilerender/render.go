package tilerender

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/golang/protobuf/proto"

	"gopnik"
	gopnik_proto "tilerender/slave/proto"
)

var FontPath string

// Renders images as Web Mercator tiles
type TileRender struct {
	command 		 []string
	cmd     		 *exec.Cmd
	writer  		 io.Writer
	reader  		 io.Reader
	executionTimeout time.Duration
}

func (t *TileRender) childLogger(cmdStderr io.Reader) {
	reader := bufio.NewReader(cmdStderr)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		log.Error("Render child error: %v", line)
	}
}

func NewTileRender(cmd []string, timeout time.Duration) (*TileRender, error) {
	t := &TileRender{
		command: cmd,
		executionTimeout: timeout,
	}

	err := t.createSubProc()

	return t, err
}

func (t *TileRender) createSubProc() error {
	if len(t.command) < 1 {
		return fmt.Errorf("Subprocess command line is to short")
	}

	t.cmd = exec.Command(t.command[0], t.command[1:]...)

	cmdStdin, err := t.cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmdStdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	t.writer = cmdStdin
	t.reader = cmdStdout

	cmdStderr, err := t.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go t.childLogger(cmdStderr)

	err = t.cmd.Start()
	if err != nil {
		return err
	}
	go t.cmd.Wait()

	err2 := t.waitReady()
	if err2 != nil {
		return err2
	}

	return err
}

func (t *TileRender) waitReady() error {
	status, err := t.readUint64()
	if err != nil {
		t.Stop()
		return err
	}
	if status != 0 {
		t.Stop()
		return fmt.Errorf(
			"Invalid child status message: %v (required 0)", status)
	}
	return nil
}

func (t *TileRender) readUint64() (uint64, error) {
	var res uint64
	err := binary.Read(t.reader, binary.LittleEndian, &res)
	if err != nil {
		return 0, fmt.Errorf("Invalid read uint64: %v", err)
	}
	return res, nil
}

func (t *TileRender) writeUint64(val uint64) error {
	writeErr := binary.Write(t.writer, binary.LittleEndian, val)
	if writeErr != nil {
		return fmt.Errorf("Write error: %v", writeErr)
	}
	return nil
}

func (t *TileRender) RenderTiles(c gopnik.TileCoord) ([]gopnik.Tile, error) {
	if t.cmd == nil {
		if err := t.createSubProc(); err != nil {
			t.Stop()
			return nil, err
		}
	}

	err := t.writeTask(c)
	if err != nil {
		t.Stop()
		return nil, err
	}

	ch := make(chan struct{})
	if t.executionTimeout > 0 {
		go func() {
			select {
			case <-time.After(t.executionTimeout):
				log.Debug("Stopping worker by timeout. TileCoord: %v", c)
				t.Stop()
				<- ch
			case <-ch:
				return
			}
		}()
		defer func(){
			ch <- struct{}{}
		}()
	}

	tiles, err := t.readAnswer()

	if err != nil {
		t.Stop()
		return nil, err
	}

	if uint64(len(tiles)) != c.Size*c.Size {
		return nil, fmt.Errorf("Invalid answer size (%v tiles)", len(tiles))
	}

	return tiles, nil
}

func (t *TileRender) writeTask(c gopnik.TileCoord) error {
	task := gopnik_proto.Task{
		X:    proto.Uint64(c.X),
		Y:    proto.Uint64(c.Y),
		Zoom: proto.Uint64(c.Zoom),
		Size: proto.Uint64(c.Size),
	}

	data, marshalErr := proto.Marshal(&task)
	if marshalErr != nil {
		return fmt.Errorf("marshaling error: ", marshalErr)
	}
	writeErr := t.writeUint64(uint64(len(data)))
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = t.writer.Write(data)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func (t *TileRender) readAnswer() ([]gopnik.Tile, error) {
	// Read size
	messageSize, readSizeErr := t.readUint64()
	if readSizeErr != nil {
		return nil, readSizeErr
	}

	// Read message buffer
	buf := make([]byte, messageSize)
	_, readErr := io.ReadFull(t.reader, buf)
	if readErr != nil {
		return nil, readErr
	}

	// Decode message
	res := &gopnik_proto.Result{}
	unmarshalErr := proto.Unmarshal(buf, res)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Convert message to internal format
	tiles := make([]gopnik.Tile, len(res.Tiles))
	for i := 0; i < len(res.Tiles); i++ {
		tiles[i].Image = res.Tiles[i].GetPng()
		col := res.Tiles[i].GetSingleColor()
		if col != nil {
			tiles[i].SingleColor = gopnik.RGBAColor{
				R: col.GetR(),
				G: col.GetG(),
				B: col.GetB(),
				A: col.GetA(),
			}
		}
	}

	return tiles, nil
}

func (t *TileRender) Stop() {
	if t.cmd != nil {
		t.cmd.Process.Kill()
	}
	t.cmd = nil
	t.reader = nil
	t.writer = nil
}
