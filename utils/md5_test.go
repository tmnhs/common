package utils_test

import (
	"github.com/tmnhs/common/utils"
	"testing"
)

func TestMD5(t *testing.T) {
	t.Log(utils.MD5("123456"))
}
