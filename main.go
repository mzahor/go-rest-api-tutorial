package main

import "os"

func main() {
	a := App{}

	a.Initialize(
		os.Getenv("USER"),
		os.Getenv("PASSWORD"),
		os.Getenv("DBNAME"),
	)

	a.Run(":8010")
}
