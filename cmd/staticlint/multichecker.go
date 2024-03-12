// Пакет multichecker включает в себя следующие анализаторы:
// exitanalyzer (проверка наличия os.Exit() в функции и файле main),
// errcheck (проверка обработки ошибок),
// analysis (стандартный пакет линтеров).
package multichecker

import (
	"github.com/knstch/shortener/cmd/staticlint/exitanalyzer"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var mychecks []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	mychecks = append(mychecks, printf.Analyzer, shadow.Analyzer, structtag.Analyzer, errcheck.Analyzer, exitanalyzer.OSExitAnalyzer)

	multichecker.Main(
		mychecks...,
	)
}
