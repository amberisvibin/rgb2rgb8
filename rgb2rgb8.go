package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/ogier/pflag"
	"github.com/wayneashleyberry/truecolor/pkg/color"
	"math"
	"os"
)

var (
	outputFile string
	verbose    bool
	version    bool
	colors     bool
)

func init() {
	pflag.StringVarP(
		&outputFile,
		"output",
		"o",
		"out.hex",
		"output file",
	)
	pflag.BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"verbose output",
	)

	pflag.BoolVarP(
		&colors,
		"colors",
		"c",
		false,
		"print colors to terminal (needs 24 bit color support)",
	)
	pflag.BoolVar(
		&version,
		"version",
		false,
		"version info",
	)
}

//converts different size bytes to float64 to compare
//high and low and return the one thats closer
func compareChannels(base, high, low byte) byte {
	absHigh := math.Abs(float64(high - base))
	absLow := math.Abs(float64(low - base))
	if absHigh < absLow {
		return high
	}
	return low
}

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s is a utility that converts RGB24 hex palette files to 8bit 3-3-2 RGB.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage:\n")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if version {
		fmt.Println("rgb2rgb8 v0.1")
		fmt.Println("(c) 2025 Amber Zeller")
		fmt.Println("Distributed under the MIT license")
		os.Exit(0)
	}

	if len(pflag.Args()) == 0 {
		fmt.Println("Error: No file specified.")
		os.Exit(1)
	}

	var inputFile = pflag.Args()[0]

	if verbose {
		fmt.Println("outputFile:", outputFile)
		fmt.Println("inputFile:", inputFile)
	}

	//read input file
	input, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Error: Cannot read input file", err)
		os.Exit(2)
	}

	defer input.Close()

	//create output file
	output, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error: Cannot create output file.", err)
		os.Exit(3)
	}

	defer output.Close()

	//scan through input
	scanner := bufio.NewScanner(input)
	var data string

	counter := 1
	for scanner.Scan() {
		line := scanner.Text()
		if verbose {
			fmt.Println()
			fmt.Println("Parsing line:", line)
		}
		lineHex, err := hex.DecodeString(line)
		if err != nil || len(lineHex) != 3 {
			fmt.Println("Error: Failed to parse line as hex.", err)
		} else {
			//separate rgb components
			red := byte(lineHex[0])
			green := byte(lineHex[1])
			blue := byte(lineHex[2])

			//and with 11100000
			red3bitLow := red & 224
			green3bitLow := green & 224
			//and with 11000000
			blue2bitLow := blue & 192

			//and with 11100000 and add 00011111
			red3bitHigh := (red & 224) + 31
			green3bitHigh := (green & 224) + 31
			//and with 11000000 and add 00111111
			blue2bitHigh := (blue & 192) + 63

			//combine color channels into bytes again
			rgb8High := []byte{red3bitHigh, green3bitHigh, blue2bitHigh}
			rgb8Low := []byte{red3bitLow, green3bitLow, blue2bitLow}

			//compare channels to find closer one to original
			redFinal := compareChannels(red, red3bitHigh, red3bitLow)
			greenFinal := compareChannels(green, green3bitHigh, green3bitLow)
			blueFinal := compareChannels(blue, blue2bitHigh, blue2bitLow)

			//combine final color channels
			rgb8Final := []byte{redFinal, greenFinal, blueFinal}

			if verbose {
				fmt.Println("Raw RGB values:", red, green, blue)
				fmt.Println("High values:", red3bitHigh, green3bitHigh, blue2bitHigh)
				fmt.Println("Low values:", red3bitLow, green3bitLow, blue2bitLow)
				fmt.Println("Final values:", redFinal, greenFinal, blueFinal)
				fmt.Println("Hex high value:", hex.EncodeToString(rgb8High))
				fmt.Println("Hex high value:", hex.EncodeToString(rgb8Low))
				fmt.Println("Hex final value:", hex.EncodeToString(rgb8Final))
			}

			//generate pretty color patterns
			if colors {
				fmt.Println("Raw  High Low  Final")
				color.Background(red, green, blue).Print("     ")
				color.Background(red3bitHigh, green3bitHigh, blue2bitHigh).Print("     ")
				color.Background(red3bitLow, green3bitLow, blue2bitLow).Print("     ")
				color.Background(redFinal, greenFinal, blueFinal).Println("     ")
			}

			data += hex.EncodeToString(rgb8Final) + "\n"
		}
		counter++
	}

	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}

	_, err = output.WriteString(data)
	if err != nil {
		fmt.Println("Error: Failed to write output file.", err)
		os.Exit(4)
	}
}
