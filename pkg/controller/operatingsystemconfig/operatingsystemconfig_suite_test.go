package operatingsystemconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOperatingSystemConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller OperatingSystemConfig Suite")
}
