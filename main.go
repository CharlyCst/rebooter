package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

const KERNEL_FILE = "/tmp/kernel.img"

var diskLock sync.Mutex

type kernelHandler struct {
	disk      string
	menuEntry string
}

func main() {
	disk := os.Args[1]
	menuEntry := os.Args[2]
	fmt.Printf("Starting rebooter server\n    Disk:       %s\n    Menu Entry: %s\n", disk, menuEntry)
	fmt.Printf("Running as %s\n", getCurrentUser())

	handler := kernelHandler{
		disk:      disk,
		menuEntry: menuEntry,
	}
	http.Handle("/", handler)

	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

func (h kernelHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing payload: kernel image not found")
		return
	}
	fmt.Println("Received a kernel image")

	// Lock disk
	diskLock.Lock()
	defer diskLock.Unlock()

	// Write kernel to a temporary file, and copy it to the target disk
	file, err := os.Create(KERNEL_FILE)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to open kernel file: %v\n", err)
		return
	}

	_, err = io.Copy(file, req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to copy kernel to file %v\n", err)
		file.Close()
		return
	}

	err = file.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to close file %v\n", err)
		return
	}

	input := fmt.Sprintf("if=%s", KERNEL_FILE)
	output := fmt.Sprintf("of=%s", h.disk)
	_, err = exec.Command("dd", input, output).Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to write kernel to disk: %v\n", err)
		return
	}

	// Configure next reboot to boot the newly installed kernel
	_, err = exec.Command("grub-reboot", h.menuEntry).Output()
	if err != nil {
		fmt.Fprintf(w, "Failed to instruct GRUB to reboot on kernel: %v\n", err)
		w.WriteHeader(500)
		return
	}

	fmt.Fprintf(w, "Kernel image installed, rebooting...\n")
	go delayedReboot(1 * time.Second)
}

func getCurrentUser() string {
	user, err := exec.Command("whoami").Output()
	if err != nil {
		return ""
	}
	return string(user)
}

func delayedReboot(d time.Duration) {
	time.Sleep(d)
	exec.Command("reboot").Output()
}
