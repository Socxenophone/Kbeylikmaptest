package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type OSM struct {
	XMLName xml.Name `xml:"osm"`
	Nodes   []Node   `xml:"node"`
	Ways    []Way    `xml:"way"`
}

type Node struct {
	XMLName xml.Name `xml:"node"`
	ID      string   `xml:"id,attr"`
	Lat     string   `xml:"lat,attr"`
	Lon     string   `xml:"lon,attr"`
}

type Way struct {
	XMLName xml.Name `xml:"way"`
	ID      string   `xml:"id,attr"`
	Nodes   []Nd     `xml:"nd"`
	Tags    []Tag    `xml:"tag"`
}

type Nd struct {
	XMLName xml.Name `xml:"nd"`
	Ref     string   `xml:"ref,attr"`
}

type Tag struct {
	XMLName xml.Name `xml:"tag"`
	Key     string   `xml:"k,attr"`
	Value   string   `xml:"v,attr"`
}

func fetchMapData() ([]byte, error) {
	bbox := "31.0,37.0,33.0,39.0"
	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/map?bbox=%s", bbox)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch map data: %s", response.Status)
	}

	return ioutil.ReadAll(response.Body)
}

func main() {
	mapData, err := fetchMapData()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	osm := OSM{}
	if err := xml.Unmarshal(mapData, &osm); err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	file, err := os.Create("map.html")
	if err != nil {
		fmt.Println("Error creating HTML file:", err)
		return
	}
	defer file.Close()

	file.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<title>Map of The Beylik of Karaman and its Surrounding Territories</title>\n")
	file.WriteString("<meta charset=\"utf-8\" />\n")
	file.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	file.WriteString("<link rel=\"stylesheet\" href=\"https://unpkg.com/leaflet/dist/leaflet.css\" />\n")
	file.WriteString("<script src=\"https://unpkg.com/leaflet/dist/leaflet.js\"></script>\n")
	file.WriteString("<style> #map { height: 100%; width: 100%; } </style>\n")
	file.WriteString("</head>\n<body>\n<div id=\"map\"></div>\n<script>\n")

	file.WriteString("var map = L.map('map').setView([38, 32], 8);\n")
	file.WriteString("L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {\n")
	file.WriteString("attribution: '&copy; <a href=\"https://www.openstreetmap.org/copyright\">OpenStreetMap</a> contributors'\n")
	file.WriteString("}).addTo(map);\n")

	for _, node := range osm.Nodes {
		lat := node.Lat
		lon := node.Lon
		file.WriteString(fmt.Sprintf("L.marker([%s, %s]).addTo(map);\n", lat, lon))
	}

	for _, way := range osm.Ways {
		if way.Tags[0].Value == "administrative" {
			file.WriteString("var latlngs = [\n")
			for _, nd := range way.Nodes {
				for _, node := range osm.Nodes {
					if nd.Ref == node.ID {
						lat := node.Lat
						lon := node.Lon
						file.WriteString(fmt.Sprintf("[%s, %s],\n", lat, lon))
					}
				}
			}
			file.WriteString("];\n")
			file.WriteString("L.polygon(latlngs, {color: 'blue'}).addTo(map);\n")
		}
	}

	file.WriteString("</script>\n</body>\n</html>")

	fmt.Println("Map HTML file generated successfully.")
}
