package main

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"app"
	"gopnik"
)

func appendLine(res []gopnik.TileCoord, line string) ([]gopnik.TileCoord, error) {
	tileUrl, err := url.Parse(line)
	if err != nil {
		return res, err
	}

	pathParts := strings.Split(tileUrl.Path[0:len(tileUrl.Path)-4], "/")

	if len(pathParts) < 3 {
		return res, errors.New("Invalid url")
	}

	z, _ := strconv.ParseUint(pathParts[len(pathParts)-3], 10, 64)
	x, _ := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	y, _ := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 64)
	tags := tileUrl.Query()["tag"]

	tile := gopnik.TileCoord{
		X:    x,
		Y:    y,
		Zoom: z,
		Tags: tags,
	}

	metaTile := app.App.Metatiler().TileToMetatile(&tile)

	for _, c := range res {
		if c.Equals(&metaTile) {
			return res, nil
		}
	}

	res = append(res, metaTile)

	return res, nil
}

func readUrlsFile(fName string) (res []gopnik.TileCoord, err error) {
	fin, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	reader := bufio.NewReader(fin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		res, err = appendLine(res, line)
		if err != nil {
			return nil, err
		}
	}
	return
}
