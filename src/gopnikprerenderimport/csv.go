package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopnik"
)

func findFieldIds(tbl [][]string) (res [4]int, err error) {
	lat, lon := [2]int{-1, -1}, [2]int{-1, -1}
	for i, fieldRaw := range tbl[0] {
		field := strings.ToLower(fieldRaw)
		if strings.Contains(field, "lat") {
			if strings.Contains(field, "min") {
				lat[0] = i
			}
			if strings.Contains(field, "max") {
				lat[1] = i
			}
		}
		if strings.Contains(field, "lon") {
			if strings.Contains(field, "min") {
				lon[0] = i
			}
			if strings.Contains(field, "max") {
				lon[1] = i
			}
		}
		if lat[0] >= 0 && lat[1] >= 0 && lon[0] >= 0 && lon[1] >= 0 {
			res[0], res[1], res[2], res[3] = lat[0], lat[1], lon[0], lon[1]
			return
		}
	}
	err = fmt.Errorf("Not enough fields: lat = %v, lon = %v", lat, lon)
	return
}

func parseBbox(data [][]string, lineNumber int, fields []int) (bbox [4]float64, err error) {
	for i, field := range fields {
		str := data[lineNumber][field]
		str = strings.Replace(str, ",", ".", -1)
		bbox[i], err = strconv.ParseFloat(str, 64)
		if err != nil {
			return
		}
	}
	return
}

func readCSVFile(fName string, zooms []uint64) (res []gopnik.TileCoord, err error) {
	fin, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer fin.Close()
	res, err = readCSV(fin, zooms)
	return
}

func readCSV(input io.Reader, zooms []uint64) (res []gopnik.TileCoord, err error) {
	reader := csv.NewReader(input)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	fields, err := findFieldIds(rows)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(rows); i++ {
		bbox, err := parseBbox(rows, i, fields[:])
		if err != nil {
			return nil, err
		}
		for _, zoom := range zooms {
			coords, err := genCoords(bbox, zoom)
			if err != nil {
				return nil, err
			}
			for _, elem := range coords {
				res = append(res, elem)
			}
		}
	}
	return
}
