package redis

import "testing"

const DefaultAddr = "127.0.0.1:6379"

func TestNewAgent(t *testing.T) {
	_, err := NewAgent(DefaultAddr)
	if err != nil{
		t.Fatal(err)
	}
}

func TestAgent_Ping(t *testing.T) {
	agent, err := NewAgent(DefaultAddr)
	if err != nil{
		t.Fatal(err)
	}

	result, err := agent.Ping()

	if result != "PONG" || err != nil{
		t.Fatal(err)
	}
}