module github.com/simcap/auditools/loginpass

go 1.12

require (
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/simcap/auditools/passwords v0.0.0-20190419165804-3049583afebb
)

replace github.com/simcap/auditools/passwords => ../passwords
