package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"log"
)

var Logger logr.Logger

func init() {
	zapLog, _ := zap.NewDevelopment()
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	Logger = zapr.NewLogger(zapLog)
}
