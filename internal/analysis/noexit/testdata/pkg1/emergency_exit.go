package main

import "os"

// emergencyExit нужна только для того чтобы проверить что линтер ничего не скажет, т.к. мы не в функции main.
func emergencyExit() {
	os.Exit(1)
}
