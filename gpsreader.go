/*
Copyright 2024 iAchieved.it LLC

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE
OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.

SPDX-License-Identifier: ISC
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pebbe/zmq4"
	"go.bug.st/serial"
)

// Helper function to parse NMEA coordinates

// TODO: Thanks ChatGPT for all your help here, but I'm pretty sure the
// decimal degrees calculation is incorrect.  I'll need to fix this.
func parseCoordinate(coord string, direction string) (float64, error) {
	var degrees int
	var minutes float64
	var err error

	result := 0.0

	// Determine if latitude (DDMM.MMMMM) or longitude (DDDMM.MMMMM)
	if len(coord) > 10 { // Longitude (DDDMM.MMMMM)
		degrees, err = strconv.Atoi(coord[:3]) // First 3 digits
		if err != nil {
			return 0, fmt.Errorf("error parsing degrees: %v", err)
		}
		minutes, err = strconv.ParseFloat(coord[3:], 64) // Remaining part
		if err != nil {
			return 0, fmt.Errorf("error parsing minutes: %v", err)
		}
	} else { // Latitude (DDMM.MMMMM)
		degrees, err = strconv.Atoi(coord[:2]) // First 2 digits

		log.Print(degrees)
		if err != nil {
			return 0, fmt.Errorf("error parsing degrees: %v", err)
		}
		minutes, err = strconv.ParseFloat(coord[2:], 64) // Remaining part
		if err != nil {
			return 0, fmt.Errorf("error parsing minutes: %v", err)
		}
	}

	result = float64(degrees) + (minutes / 60)

	if direction == "S" || direction == "W" {
		result = -result
	}

	return result, nil

}

func parseGPGLL(sentence string) (bool, float64, float64) {
	// Split the sentence into fields
	fields := strings.Split(sentence, ",")

	// GPGLL Format: $GPGLL,lat,N,lon,W,time,status,checksum
	if len(fields) < 7 {
		log.Printf("Malformed GPGLL sentence: %s", sentence)
		return false, 0, 0
	}

	lat := fields[1]
	latDir := fields[2]
	lon := fields[3]
	lonDir := fields[4]
	status := fields[6]

	// Check if the status indicates valid data ('A' means valid, 'V' means invalid)
	if status != "A" {
		log.Printf("Invalid GPGLL sentence: %s", sentence)
		return false, 0, 0
	}

	// Convert latitude and longitude to float64
	latitude, err := parseCoordinate(lat, latDir)
	if err != nil {
		log.Printf("Invalid latitude: %s, Error: %v", lat, err)
		return false, 0, 0
	}

	longitude, err := parseCoordinate(lon, lonDir)
	if err != nil {
		log.Printf("Invalid longitude: %s, Error: %v", lon, err)
		return false, 0, 0
	}

	return true, latitude, longitude
}

func main() {
	fmt.Println("Hello, World!")

	publisher, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	err = publisher.Bind("tcp://*:11205")
	if err != nil {
		log.Fatal(err)
	}

	mode := &serial.Mode{
		BaudRate: 115200,
	}

	// For the Mac
	// port, err := serial.Open("/dev/cu.usbmodem11101", mode)

	// For the Pi
	serialDevices := []string{"/dev/ttyACM0",
		"/dev/ttyACM1",
		"/dev/ttyACM2",
		"/dev/ttyACM3"}

	var port serial.Port
	for _, device := range serialDevices {
		port, err = serial.Open(device, mode)
		if err != nil {
			log.Printf("Error opening serial port %s: (%v)\n", device, err)
		} else {
			defer port.Close()
			break
		}
	}

	reader := bufio.NewReader(port)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from serial port: %v", err)
			continue
		}

		// Trim whitespace and check if the line contains GPGLL
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "$GPGLL") {
			fmt.Printf("GPGLL: %s\n", line)
			valid, lat, lng := parseGPGLL(line)
			if valid {
				fmt.Printf("Latitude: %f, Longitude: %f\n", lat, lng)
				publisher.Send(fmt.Sprintf("%f/%f", lat, lng), 0)
			}
		}
	}

}
