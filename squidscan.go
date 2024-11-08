package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FIRST TODO: concurrency & help menu.
// TODO: add support for proxy authentication: https://wiki.squid-cache.org/Features/Authentication
// TODO: add udp & ipv6 support (optional).

type settings struct {
	minPort   int
	maxPort   int
	proxyAddr string
	proxyPort int
}

func makePortRange(min int, max int) []int {

	var s []int
	for i := min; i <= max; i++ {
		s = append(s, i)
	}
	return s
}

func portInput(msg string) (int, int) {

	fmt.Print(msg)
	var inp string
	fmt.Scan(&inp)

	before, after, _ := strings.Cut(inp, "-")

	minPort, err := strconv.Atoi(before)
	if err != nil {
		log.Println("Invalid input: You must enter integers for the port range.")
		return portInput(msg)
	}

	maxPort, err := strconv.Atoi(after)

	if err != nil {
		log.Println("Invalid input: You must enter integers for the port range.")
		return portInput(msg)
	}

	if minPort <= 0 || maxPort <= 0 || minPort > maxPort || maxPort > 65535 {
		log.Println("Invalid port range: Please enter valid ports (1-65535) and ensure min < max.")
		return portInput(msg)
	}

	return minPort, maxPort
}

func settingsInput(msg string) (string, int) {

	fmt.Print(msg)
	var settings string
	fmt.Scan(&settings)

	ip, portTmp, _ := strings.Cut(settings, ":")
	port, err := strconv.Atoi(portTmp)

	if ip == "localhost" {
		ip = "127.0.0.1"
	}

	if err != nil {
		log.Println("Invalid input: You must enter an integer for the port.")
		return settingsInput(msg)
	}

	ipMatch, _ := regexp.MatchString(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`, ip)
	if !ipMatch {
		log.Println("Invalid input: Please enter a valid IP address.")
		return settingsInput(msg)
	}

	if port <= 0 || port > 65535 {
		log.Println("Invalid port: Please enter a valid port number (1-65535).")
		return settingsInput(msg)
	}

	return ip, port
}

func confirmationPrompt(settings *settings) bool {

	var choice string
	fmt.Printf("Proxy Address: %v\n", settings.proxyAddr)
	fmt.Printf("Proxy Port: %v\n", settings.proxyPort)
	fmt.Printf("Port range: %v-%v\n", settings.minPort, settings.maxPort)
	fmt.Print("Do these settings look OK to you? <Y/N>: ")
	fmt.Scan(&choice)

	if choice == "Y" || choice == "y" {
		return true
	}

	return false
}

func settingsInit() settings {

	var settings settings
	var minPort, maxPort int

	ip, pPort := settingsInput("Enter your proxy settings <ip:port>: ")
	minPort, maxPort = portInput("Enter your port range <1-65535>: ")

	settings.proxyAddr = ip
	settings.proxyPort = pPort
	settings.minPort = minPort
	settings.maxPort = maxPort

	return settings
}

func squidScan(proxy string, targetPort int) int {

	conn, err := net.Dial("tcp4", proxy)

	if err != nil {
		log.Println("Error communicating with the proxy server: %v", err)
		os.Exit(0)
	}

	defer conn.Close()

	target := "127.0.0.1"

	var req = fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\n\r\n", target, targetPort, target, targetPort)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println("Error communicating with the local machine: %v", err)
	}

	var response [1024]byte


	n, err := conn.Read(response[:])
	if err != nil {
		log.Printf("Error communicating with the local machine: %v", err)
	}


	if string(response[:n])[:12] != "HTTP/1.1 200" {
		return 0
	}

	return targetPort

}

func squidder(settings *settings) {

	proxyCombined := settings.proxyAddr + ":" + strconv.Itoa(settings.proxyPort)

	ports := makePortRange(settings.minPort, settings.maxPort)

	for _, port := range ports {
		valid := squidScan(proxyCombined, port)
		if valid != 0 {
			log.Printf("Open port %d", valid)
		}
	}

}

func main() {

	var settings settings

	fmt.Println("Welcome to squidscan!")
	settings = settingsInit()

	ans := confirmationPrompt(&settings)

	for !ans {
		fmt.Println("Restarting...")
		settings = settingsInit()
		ans = confirmationPrompt(&settings)
	}

	fmt.Println("Well then let's get started :)")
	squidder(&settings)
}
