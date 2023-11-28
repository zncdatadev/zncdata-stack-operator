package controller

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	//准备工作
	fmt.Println("start prepare")
	_ = godotenv.Load("../../.env")
	exitCode := m.Run()
	// 结束工作
	fmt.Println("end prepare")
	os.Exit(exitCode)
	//清理工作
	fmt.Println("prepare to clean")

	os.Exit(exitCode)
}
