package goop

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Goop", func() {
	Describe("vendorDir()", func() {

		It("Utilizes 'GOOP_VENDOR_DIR' environ variable when defined", func() {
			os.Setenv("GOOP_VENDOR_DIR", "fake-goop-vendor-dir")
			dir := NewGoop("./", os.Stdin, os.Stdout, os.Stderr).vendorDir()
			Expect("fake-goop-vendor-dir").To(Equal(dir))
		})

		It("Defaults to '.vendors' in the base directory", func() {
			os.Unsetenv("GOOP_VENDOR_DIR")
			dir := NewGoop("/", os.Stdin, os.Stdout, os.Stderr).vendorDir()
			Expect("/.vendor").To(Equal(dir))
		})

	})
})

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "goop")
}
