module github.com/smartwalle/rx

require (
	github.com/smartwalle/nsync v0.0.0
)

replace (
	github.com/smartwalle/nsync => ../nsync
)

go 1.18
