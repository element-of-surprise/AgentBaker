// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package parser

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"testing"
)

var _ = BeforeSuite(func() {
})

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Parser Suite", []Reporter{junitReporter})
}
