package parsertest

import (
	"bytes"
	"testing"

	"github.com/nitrous-io/goop/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "parser")
}

var _ = Describe("parser", func() {
	Describe("Parse()", func() {
		var (
			deps []*parser.Dependency
			err  error
		)

		Context("empty Goopfile", func() {
			BeforeEach(func() {
				deps, err = parser.Parse(bytes.NewBufferString(""))
			})

			It("returns an empty slice", func() {
				Expect(err).To(BeNil())
				Expect(deps).NotTo(BeNil())
				Expect(deps).To(HaveLen(0))
			})
		})

		Context("one entry", func() {
			Context("with no revision specified", func() {
				BeforeEach(func() {
					deps, err = parser.Parse(bytes.NewBufferString(`
						github.com/nitrous-io/goop
					`))
				})

				It("parses and returns a slice containing one dependency item", func() {
					Expect(err).To(BeNil())
					Expect(deps).To(HaveLen(1))
					Expect(deps[0]).To(Equal(&parser.Dependency{Pkg: "github.com/nitrous-io/goop", Rev: ""}))
				})
			})

			Context("with revision specified", func() {
				BeforeEach(func() {
					deps, err = parser.Parse(bytes.NewBufferString(`
						github.com/nitrous-io/goop #09f0feb1b103933bd9985f0a85e01eeaad8d75c8
					`))
				})

				It("parses and returns a slice containing one dependency item", func() {
					Expect(err).To(BeNil())
					Expect(deps).To(HaveLen(1))
					Expect(deps[0]).To(Equal(&parser.Dependency{Pkg: "github.com/nitrous-io/goop", Rev: "09f0feb1b103933bd9985f0a85e01eeaad8d75c8"}))
				})
			})

			Context("with unparseable garbage", func() {
				BeforeEach(func() {
					deps, err = parser.Parse(bytes.NewBufferString(`
						github.com/nitrous-io/goop (*@#&!@(*#)@$F@sdgu8$!
					`))
				})

				It("fails and returns parse error", func() {
					Expect(err).NotTo(BeNil())
					Expect(deps).To(BeNil())
				})
			})

			Context("with only one comment", func() {
				BeforeEach(func() {
					deps, err = parser.Parse(bytes.NewBufferString(`
						//This is a comment.
					`))					
				})

				It("returns an empty slice", func() {
					Expect(err).To(BeNil())
					Expect(deps).NotTo(BeNil())
					Expect(deps).To(HaveLen(0))
				})
			})
		})

		Context("multiple entries with comment", func() {
			BeforeEach(func() {
				deps, err = parser.Parse(bytes.NewBufferString(`
					//This is a comment.
					github.com/nitrous-io/goop #09f0feb1b103933bd9985f0a85e01eeaad8d75c8

					github.com/gorilla/mux
					github.com/gorilla/context #14f550f51af52180c2eefed15e5fd18d63c0a64a
				`))
			})

			It("parses and returns a slice containing multiple dependency items", func() {
				Expect(err).To(BeNil())
				Expect(deps).To(HaveLen(3))
				Expect(deps[0]).To(Equal(&parser.Dependency{Pkg: "github.com/nitrous-io/goop", Rev: "09f0feb1b103933bd9985f0a85e01eeaad8d75c8"}))
				Expect(deps[1]).To(Equal(&parser.Dependency{Pkg: "github.com/gorilla/mux", Rev: ""}))
				Expect(deps[2]).To(Equal(&parser.Dependency{Pkg: "github.com/gorilla/context", Rev: "14f550f51af52180c2eefed15e5fd18d63c0a64a"}))
			})
		})
	})
})
