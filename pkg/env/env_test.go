package env_test

import (
	"os"
	"testing"

	"github.com/nitrous-io/goop/pkg/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "env")
}

var _ = Describe("env", func() {
	var e env.Env

	BeforeEach(func() {
		os.Setenv("_GOOP_ENV_TEST_FOO", "foo")
		os.Setenv("_GOOP_ENV_TEST_BAR", "bar=bar bar")
		os.Setenv("_GOOP_ENV_TEST_EMPTY", "")

		e = env.NewEnv()
	})

	AfterEach(func() {
		os.Setenv("_GOOP_ENV_TEST_FOO", "")
		os.Setenv("_GOOP_ENV_TEST_BAR", "")
		os.Setenv("_GOOP_ENV_TEST_EMPTY", "")
	})

	Describe("NewEnv()", func() {
		It("returns a new env map using current env vars", func() {
			Expect(e["_GOOP_ENV_TEST_FOO"]).To(Equal("foo"))
			Expect(e["_GOOP_ENV_TEST_BAR"]).To(Equal("bar=bar bar"))
			Expect(e["_GOOP_ENV_TEST_EMPTY"]).To(BeEmpty())
		})
	})

	Describe("Strings()", func() {
		It("returns a copy of strings representing the env, in the form key=value", func() {
			s := e.Strings()
			Expect(s).To(ContainElement("_GOOP_ENV_TEST_FOO=foo"))
			Expect(s).To(ContainElement("_GOOP_ENV_TEST_BAR=bar=bar bar"))
			Expect(s).To(ContainElement("_GOOP_ENV_TEST_EMPTY="))
		})
	})

	Describe("Prepend()", func() {
		Context("when a given key has a value", func() {
			It("prepends new value to the existing value", func() {
				e.Prepend("_GOOP_ENV_TEST_FOO", "lol")
				Expect(e["_GOOP_ENV_TEST_FOO"]).To(Equal("lol:foo"))
				e.Prepend("_GOOP_ENV_TEST_FOO", "hello")
				Expect(e["_GOOP_ENV_TEST_FOO"]).To(Equal("hello:lol:foo"))
			})
		})

		Context("when a given key is empty", func() {
			It("sets new value", func() {
				e.Prepend("_GOOP_ENV_TEST_EMPTY", "foo")
				Expect(e["_GOOP_ENV_TEST_EMPTY"]).To(Equal("foo"))
				e.Prepend("_GOOP_ENV_TEST_EMPTY", "lol")
				Expect(e["_GOOP_ENV_TEST_EMPTY"]).To(Equal("lol:foo"))
			})
		})

		Context("when a given key does not exist", func() {
			BeforeEach(func() {
				delete(e, "_GOOP_ENV_TEST_EMPTY")
			})

			It("sets new value", func() {
				e.Prepend("_GOOP_ENV_TEST_EMPTY", "foo")
				Expect(e["_GOOP_ENV_TEST_EMPTY"]).To(Equal("foo"))
				e.Prepend("_GOOP_ENV_TEST_EMPTY", "lol")
				Expect(e["_GOOP_ENV_TEST_EMPTY"]).To(Equal("lol:foo"))
			})
		})
	})
})
