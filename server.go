package main

import (
	"fmt"
	"log"
	"net"
	"io"
	"math"
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/credentials"

	pb "github.com/chukmunnlee/grpc_calculator/messages"
	ptypes "github.com/golang/protobuf/ptypes"
)

const TCP = "tcp";
const INTERFACE = "0.0.0.0:50051";

type server struct {
	pb.UnimplementedCalculatorServiceServer
}

func (*server) Calculate(stream pb.CalculatorService_CalculateServer) error {
	log.Println("New calculation")

	result := float32(0.0)
	count := uint32(0)
	id := ""

	for {
		req, err := stream.Recv()
		if io.EOF == err {
			resp := &pb.CalculateResponse{
				Id: id,
				Result: result,
				Operations: count,
				Timestamp: ptypes.TimestampNow(),
			}
			return stream.SendAndClose(resp)
		};
		if nil != err {
			log.Printf("Error: %v\n", err)
			return err
		}

		id = req.GetId()
		operation := req.GetOperation()
		operand := operation.GetOperand()
		operator := operation.GetOperator().String()
		count++

		log.Printf("[%s-%d] operator: %s, operand: %f\n", id, count, operator, operand)

		switch (operator) {
			case "SET":
				result = operand;
				break;

			case "ADD":
				result += operand;
				break;

			case "SUB":
				result -= operand;
				break;

			case "DIV":
				result /= operand;
				break;

			case "MUL":
				result *= operand;
				break;

			case "MOD":
				result = float32(math.Mod(float64(result), float64(operand)))
				break;

			default:
				log.Printf("Skipping unknown operation: %s %f\n", operator, operand)
		}
	}

	return nil
}

type CliOptions struct {
	tls bool
}

func parseCommandline() []grpc.ServerOption {
	opts := [] grpc.ServerOption{}

	tls := flag.Bool("tls", false, "Enable TLS")
	certFile := flag.String("certFile", "", "Server cert")
	keyFile := flag.String("keyFile", "", "Server key file")

	flag.Parse()

	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if nil != err {
			log.Panicf("Cannot load cert/key files: %v\n", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	return opts
}

func main() {
	fmt.Printf("Calculator server v%d\n", 1)

	lis, err  := net.Listen(TCP, INTERFACE)
	if nil != err {
		log.Fatalf("Cannot listen on %s\n", INTERFACE)
	}

	opts := parseCommandline()

	fmt.Printf("cli: %v\n", opts)

	s := grpc.NewServer(opts...)

	pb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)

	log.Printf("Starting CalculatorService")
	if err := s.Serve(lis); nil != err {
		log.Fatalf("Cannot start CalculatorService: %v\n", err);
	}
}
