package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"runtime"
	"strconv"

	utils "github.com/krnblni/UnaryGoClientStreamGRPC/proto/go"
	"google.golang.org/grpc"
)

type utilsServer struct{}

func (us *utilsServer) UploadFileAndGetSize(stream utils.Utils_UploadFileAndGetSizeServer) error {
	_, filename, _, _ := runtime.Caller(0)
	// doing this so that from where ever the user runs this file,
	// the temp file will be stored in this directory only
	currentPath := filepath.Dir(filename)
	tempFile, err := ioutil.TempFile(currentPath + "/files", "recieved-")
	if err != nil {
		fmt.Println("Unable to create tempfile - ", err)
		return err
	}
	fmt.Println("Created File: ", tempFile.Name())
	for {
		fileSegment, err := stream.Recv()
		if err == io.EOF {
			// means whole file sent by client
			// return size of file - string

			// get file stat
			fileStat, err := tempFile.Stat()
			if err != nil {
				return err
			}
			sizeString := strconv.Itoa(int(fileStat.Size()))

			// close the file
			if err := tempFile.Close(); err != nil {
				log.Fatal("Error closing file...", err)
			}

			return stream.SendAndClose(&utils.FileSize{Size: sizeString})
		}
		if err != nil {
			// some other error occured
			return err
		}
		// add file segment to tempfile
		fileSegmentData := fileSegment.GetFileSegmentData()
		if _, err := tempFile.Write(fileSegmentData); err != nil {
			log.Fatal("Error writing to file...", err)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Unable to create a listener - ", err)
		return
	}

	utilsGrpcServer := grpc.NewServer()

	utils.RegisterUtilsServer(utilsGrpcServer, &utilsServer{})

	if err := utilsGrpcServer.Serve(listener); err != nil {
		fmt.Println("Unable to start the server - ", err)
		return
	}
}
