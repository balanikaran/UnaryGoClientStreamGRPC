package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	utils "github.com/krnblni/UnaryGoClientStreamGRPC/proto/go"
	"google.golang.org/grpc"
)

const bufferSize = 128 //bytes

func main() {

	connection, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		fmt.Println("Unable to create a new connection...", err)
	}

	client := utils.NewUtilsClient(connection)
	stream, err := client.UploadFileAndGetSize(context.Background())
	if err != nil {
		fmt.Println("Cannot get client stream... ", err)
	}

	// reading file and sending to server
	// 128 bytes chunck wise
	// getting the current path for file
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Unable to get file path.")
	}

	currentPath := filepath.Dir(thisFile)
	fmt.Println(currentPath)

	file, err := os.Open(currentPath + "/filetosend")
	if err != nil {
		fmt.Println("Unable to open file - ", err)
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)

	for {
		numberOfBytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Unable to read file - ", err)
		}
		stream.Send(&utils.FileSegment{FileSegmentData: buffer[:numberOfBytesRead]})
		// fmt.Println(string(buffer[:numberOfBytesRead]))
	}

	fileSize, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("Unable to get response from server after sending and closing stream...", err)
		return
	}

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println("Unable to get file stats - ", err)
		return 
	}
	fmt.Println("Size of file sent from client: ", fileStat.Size())
	fmt.Println("Size of recieved at server: ", fileSize)

}
