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
func parseCoordinate(coord, direction string) (float64, error) {
	// Example: Convert "4916.45" (49°16.45') into decimal degrees
	if coord == "" || direction == "" {
		return 0, fmt.Errorf("empty coordinate or direction")
	}

	// Parse degrees and minutes from the NMEA format
	degrees, err := strconv.ParseFloat(coord[:len(coord)-7], 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(coord[len(coord)-7:], 64)
	if err != nil {
		return 0, err
	}

	// Calculate decimal degrees
	decimalDegrees := (degrees + minutes/60) / 10

	// Adjust for direction
	if direction == "S" || direction == "W" {
		decimalDegrees = -decimalDegrees
	}

	return decimalDegrees, nil
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

	port, err := serial.Open("/dev/cu.usbmodem11101", mode)
	if err != nil {
		log.Println("Error opening serial port")
		log.Fatal(err)
	}
	defer port.Close()

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