package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"
	"context"
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/google/uuid"
	pb "github.com/chukmunnlee/grpc_calculator/messages"
)

const SERVICE_ENDPOINT = "localhost:50051"

func parseCommandLine() []grpc.DialOption {
	opts := []grpc.DialOption{};

	tls := flag.Bool("tls", false, "Enable TLS")
	caCert := flag.String("caCert", "", "CA root cert")

	flag.Parse()
	if *tls {
		creds, err := credentials.NewClientTLSFromFile(*caCert, "")
		if nil != err {
			log.Panicf("Cannot load CA cert: %v\n", err)
		}
		fmt.Println("Enabling SSL")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		fmt.Println("No SSL enabled")
		opts = append(opts, grpc.WithInsecure())
	}

	return opts
}

func findArgs(args []string) int {
	idx := 1

	for i, v := range args {
		if "--" == v {
			return i + 1
		}
	}

	return idx
}

func main() {

	opts := parseCommandLine()

	conn, err := grpc.Dial(SERVICE_ENDPOINT, opts...)
	if nil != err {
		log.Fatalf("Cannot connect to service: %v\n", err)
	}
	defer conn.Close()

	c := pb.NewCalculatorServiceClient(conn)

	stream, err := c.Calculate(context.Background())
	if nil != err {
		log.Fatalf("Cannot invoke Calculate(): %v\n", err)
	}

	uid, err := uuid.NewRandom()
	if nil != err {
		log.Fatalf("Cannot create uuid: %v\n", err)
	}
	id := uid.String()[:8]

	idx := findArgs(os.Args)
	seq := uint32(0)

	for idx < len(os.Args) {
		operator := strings.ToUpper(os.Args[idx])
		idx++
		operand, err := strconv.ParseFloat(os.Args[idx], 32)
		if nil != err {
			log.Fatalf("Not a number: %s\n", os.Args[idx])
		}
		idx++
		seq++
		fmt.Printf("[%s-%d] operator: %s, operand: %f\n", id, seq, operator, operand)

		oper := &pb.Operation {
			Operand: float32(operand),
			Operator: pb.Operation_Operator(pb.Operation_Operator_value[operator]),
		}
		req := &pb.CalculateRequest {
			Id: id,
			Seq: seq,
			Operation: oper,
		}
		if err := stream.Send(req); nil != err {
			log.Fatalf("Send error: %v\n", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if nil != err {
		log.Fatalf("Calculation error: %v\n", err)
	}

	fmt.Printf("Response: %v\n", resp)

}
