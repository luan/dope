package fetcher_test

import (
	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/luan/idope/fetcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetcher", func() {
	It("builds", func() {
		bbsClient := &fake_bbs.FakeClient{}
		fDawg := fetcher.NewFetcher(bbsClient)
		Expect(fDawg).NotTo(BeNil())
	})
})
