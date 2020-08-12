package kafka

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func checkClose(t *testing.T, c io.Closer) {
	require.NoError(t, c.Close())
}

func logger() logrus.FieldLogger {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)

	return log
}
