package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("kubectl", "exec", "-n", "staging", "-ti", "swarm-public-0", "--", "cat", "/root/.ethereum/swarm/nodekey")
	res, err := cmd.Output()
	fmt.Println(string(res))
}
