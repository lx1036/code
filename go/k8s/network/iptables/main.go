package main

func main() {
	err := iptables.SetupIPForward()
	if err != nil {
		panic(err)
	}
}
