package main

import "testing"

func TestParseHamQSLUsesProviderRatings(t *testing.T) {
	body := []byte(`<solar><solardata><updated>20 Nov 2026 0818 GMT</updated>
	<solarflux>107</solarflux><aindex> 14</aindex><kindex>2</kindex><fof2></fof2>
	<calculatedconditions><band name="80m-40m" time="night">Good</band>
	<band name="80m-40m" time="day">Fair</band><band name="12m-10m" time="day">Poor</band></calculatedconditions>
	<calculatedvhfconditions><phenomenon name="E-Skip" location="europe_6m">Band Closed</phenomenon></calculatedvhfconditions>
	</solardata></solar>`)
	report, err := parseHamQSL(body)
	if err != nil {
		t.Fatal(err)
	}
	if report.UpdatedAt.Format("2006-01-02 15:04 MST") != "2026-11-20 08:18 UTC" ||
		report.SFI != "107" || report.A != "14" || report.K != "2" || report.FOF2 != "" ||
		len(report.HF) != 2 || report.HF[0] != (hfCondition{"80m-40m", "Fair", "Good"}) ||
		len(report.VHF) != 1 || report.VHF[0] != (vhfCondition{"E-Skip", "europe_6m", "Band Closed"}) {
		t.Fatalf("unexpected solar report: %+v", report)
	}
	if _, err := parseHamQSL([]byte("<html>outage</html>")); err == nil {
		t.Fatal("non-XML result accepted as solar data")
	}
}
