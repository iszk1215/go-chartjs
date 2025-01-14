// Package chartjs simplifies making chartjs.org plots in go.
package chartjs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"

	"github.com/iszk1215/go-chartjs/types"
)

var True = types.True
var False = types.False

var chartTypes = [...]string{
	"line",
	"bar",
	"bubble",
}

type chartType int

func (c chartType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + chartTypes[c] + `"`), nil
}

const (
	// Line is a "line" plot
	Line chartType = iota
	// Bar is a "bar" plot
	Bar
	// Bubble is a "bubble" plot
	Bubble
)

type interpMode int

const (
	_ interpMode = iota
	InterpMonotone
	InterpDefault
)

var interpModes = [...]string{
	"",
	"monotone",
	"default",
}

func (m interpMode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + interpModes[m] + `"`), nil
}

// XFloatFormat determines how many decimal places are sent in the JSON for X values.
var XFloatFormat = "%.2f"

// YFloatFormat determines how many decimal places are sent in the JSON for Y values.
var YFloatFormat = "%.2f"

// Values dictates the interface of data to be plotted.
type Values interface {
	// X-axis values. If only these are specified then it must be a Bar plot.
	Xs() []float64
	// Optional Y values.
	Ys() []float64
	// Rs are used to size points for chartType `Bubble`
	Rs() []float64
}

func marshalValuesJSON(v Values, xformat, yformat string) ([]byte, error) {
	xs, ys, rs := v.Xs(), v.Ys(), v.Rs()
	if len(xs) == 0 {
		if len(rs) != 0 {
			return nil, fmt.Errorf("chart: bad format of Values data")
		}
		xs = ys[:len(ys)]
		ys = nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, 8*len(xs)))
	buf.WriteRune('[')
	if len(rs) > 0 {
		if len(xs) != len(ys) || len(xs) != len(rs) {
			return nil, fmt.Errorf("chart: bad format of Values. All axes must be of the same length")
		}
		var err error
		for i, x := range xs {
			if i > 0 {
				buf.WriteRune(',')
			}
			y, r := ys[i], rs[i]
			if math.IsNaN(y) {
				_, err = buf.WriteString(fmt.Sprintf(("{\"x\":" + xformat + ",\"y\": null,\"r\":" + yformat + "}"), x, r))
			} else {
				_, err = buf.WriteString(fmt.Sprintf(("{\"x\":" + xformat + ",\"y\":" + yformat + ",\"r\":" + yformat + "}"), x, y, r))
			}
			if err != nil {
				return nil, err
			}
		}
	} else if len(ys) > 0 {
		if len(xs) != len(ys) {
			return nil, fmt.Errorf("chart: bad format of Values. X and Y must be of the same length")
		}
		var err error
		for i, x := range xs {
			if i > 0 {
				buf.WriteRune(',')
			}
			y := ys[i]
			if math.IsNaN(y) {
				_, err = buf.WriteString(fmt.Sprintf(("{\"x\":" + xformat + ",\"y\": null }"), x))
			} else {
				_, err = buf.WriteString(fmt.Sprintf(("{\"x\":" + xformat + ",\"y\":" + yformat + "}"), x, y))
			}
			if err != nil {
				return nil, err
			}
		}

	} else {
		for i, x := range xs {
			if i > 0 {
				buf.WriteRune(',')
			}
			_, err := buf.WriteString(fmt.Sprintf(xformat, x))
			if err != nil {
				return nil, err
			}
		}
	}

	buf.WriteRune(']')
	return buf.Bytes(), nil
}

// shape indicates the type of marker used for plotting.
type shape int

var shapes = []string{
	"",
	"circle",
	"triangle",
	"rect",
	"rectRot",
	"cross",
	"crossRot",
	"star",
	"line",
	"dash",
}

const (
	empty = iota
	Circle
	Triangle
	Rect
	RectRot
	Cross
	CrossRot
	Star
	LinePoint
	Dash
)

func (s shape) MarshalJSON() ([]byte, error) {
	return []byte(`"` + shapes[s] + `"`), nil
}

// Dataset wraps the "dataset" JSON
type Dataset struct {
	Data            interface{} `json:"-"`
	Type            chartType   `json:"type,omitempty"`
	BackgroundColor *types.RGBA `json:"backgroundColor,omitempty"`
	// BorderColor is the color of the line.
	BorderColor *types.RGBA `json:"borderColor,omitempty"`
	// BorderWidth is the width of the line.
	BorderWidth float64 `json:"borderWidth"`

	// Label indicates the name of the dataset to be shown in the legend.
	Label string     `json:"label,omitempty"`
	Fill  types.Bool `json:"fill,omitempty"`

	// SteppedLine of true means dont interpolate and ignore line tension.
	SteppedLine            types.Bool  `json:"steppedLine,omitempty"`
	LineTension            float64     `json:"lineTension"`
	CubicInterpolationMode interpMode  `json:"cubicInterpolationMode,omitempty"`
	PointBackgroundColor   *types.RGBA `json:"pointBackgroundColor,omitempty"`
	PointBorderColor       *types.RGBA `json:"pointBorderColor,omitempty"`
	PointBorderWidth       float64     `json:"pointBorderWidth"`
	PointRadius            float64     `json:"pointRadius"`
	PointHitRadius         float64     `json:"pointHitRadius"`
	PointHoverRadius       float64     `json:"pointHoverRadius"`
	PointHoverBorderColor  *types.RGBA `json:"pointHoverBorderColor,omitempty"`
	PointHoverBorderWidth  float64     `json:"pointHoverBorderWidth"`
	PointStyle             shape       `json:"pointStyle,omitempty"`

	ShowLine types.Bool `json:"showLine,omitempty"`
	SpanGaps types.Bool `json:"spanGaps,omitempty"`

	// Axis ID that matches the ID on the Axis where this dataset is to be drawn.
	XAxisID string `json:"xAxisID,omitempty"`
	YAxisID string `json:"yAxisID,omitempty"`

	// set the formatter for the data, e.g. "%.2f"
	// these are not exported in the json, just used to determine the decimals of precision to show
	XFloatFormat string `json:"-"`
	YFloatFormat string `json:"-"`
}

// MarshalJSON implements json.Marshaler interface.
func (d Dataset) MarshalJSON() ([]byte, error) {
	xf, yf := d.XFloatFormat, d.YFloatFormat
	if xf == "" {
		xf = XFloatFormat
	}
	if yf == "" {
		yf = YFloatFormat
	}

	var err error
	var o []byte
	if m, ok := d.Data.(json.Marshaler); ok {
		o, err = m.MarshalJSON()
	} else if v, ok := d.Data.(Values); ok {
		o, err = marshalValuesJSON(v, xf, yf)
	}
	if err != nil {
		return nil, err
	}
	// avoid recursion by creating an alias.
	type alias Dataset
	buf, err := json.Marshal(alias(d))
	if err != nil {
		return nil, err
	}
	// replace '}' with ',' to continue struct
	if len(buf) > 0 {
		buf[len(buf)-1] = ','
	}
	buf = append(buf, []byte(`"data":`)...)
	buf = append(buf, o...)
	buf = append(buf, '}')
	return buf, nil
}

// Data wraps the "data" JSON
type Data struct {
	Datasets []Dataset `json:"datasets"`
	Labels   []string  `json:"labels"`
}

type axisType int

var axisTypes = []string{
	"category",
	"linear",
	"logarithmic",
	"time",
	"radialLinear",
}

const (
	// Category is a categorical axis (this is the default),
	// used for bar plots.
	Category axisType = iota
	// Linear axis should be use for scatter plots.
	Linear
	// Log axis
	Log
	// Time axis
	Time
	// Radial axis
	Radial
)

func (t axisType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + axisTypes[t] + "\""), nil
}

type axisPosition int

const (
	// Bottom puts the axis on the bottom (used for Y-axis)
	Bottom axisPosition = iota + 1
	// Top puts the axis on the bottom (used for Y-axis)
	Top
	// Left puts the axis on the bottom (used for X-axis)
	Left
	// Right puts the axis on the bottom (used for X-axis)
	Right
)

var axisPositions = []string{
	"",
	"bottom",
	"top",
	"left",
	"right",
}

func (p axisPosition) MarshalJSON() ([]byte, error) {
	return []byte(`"` + axisPositions[p] + `"`), nil
}

type AxisTitle struct {
	Display bool   `json:"display,omitempty"`
	Text    string `json:"text,omitempty"`
}

// Axis corresponds to 'scale' in chart.js lingo.
type Axis struct {
	Type      axisType     `json:"type"`
	Position  axisPosition `json:"position,omitempty"`
	Label     string       `json:"label,omitempty"`
	ID        string       `json:"-"`
	GridLines types.Bool   `json:"gridLine,omitempty"`
	Stacked   types.Bool   `json:"stacked,omitempty"`

	// Bool differentiates between false and empty by use of pointer.
	Display    types.Bool  `json:"display,omitempty"`
	ScaleLabel *ScaleLabel `json:"scaleLabel,omitempty"`
	Tick       *Tick       `json:"ticks,omitempty"`

	Title AxisTitle `json:"title,omitempty"`
}

// Tick lets us set the range of the data.
type Tick struct {
	Min         float64    `json:"min,omitempty"`
	Max         float64    `json:"max,omitempty"`
	BeginAtZero types.Bool `json:"beginAtZero,omitempty"`
	// TODO: add additional options from: tick options.
}

// ScaleLabel corresponds to scale title.
// Display: True must be specified for this to be shown.
type ScaleLabel struct {
	Display     types.Bool  `json:"display,omitempty"`
	LabelString string      `json:"labelString,omitempty"`
	FontColor   *types.RGBA `json:"fontColor,omitempty"`
	FontFamily  string      `json:"fontFamily,omitempty"`
	FontSize    int         `json:"fontSize,omitempty"`
	FontStyle   string      `json:"fontStyle,omitempty"`
}

// Option wraps the chartjs "option"
type Option struct {
	Responsive          types.Bool `json:"responsive,omitempty"`
	MaintainAspectRatio types.Bool `json:"maintainAspectRatio,omitempty"`
	Title               *Title     `json:"title,omitempty"`
}

// Title is the Options title
type Title struct {
	Display types.Bool `json:"display,omitempty"`
	Text    string     `json:"text,omitempty"`
}

type Animation struct {
	Duration int `json:"duration"`
}

// Options wraps the chartjs "options"
type Options struct {
	Option
	Scales    map[string]Axis              `json:"scales,omitempty"`
	Legend    *Legend                      `json:"legend,omitempty"`
	Tooltip   *Tooltip                     `json:"tooltips,omitempty"`
	Animation Animation                    `json:"animation,omitempty"`
	Plugins   map[string]map[string]string `json:"plugins,omitempty"`
}

// Tooltip wraps chartjs "tooltips".
// TODO: figure out how to make this work.
type Tooltip struct {
	Enabled   types.Bool `json:"enabled,omitempty"`
	Intersect types.Bool `json:"intersect,omitempty"`
	// TODO: make mode typed by Interaction modes.
	Mode   string         `json:"mode,omitempty"`
	Custom template.JSStr `json:"custom,omitempty"`
}

type Legend struct {
	Display types.Bool `json:"display,omitempty"`
}

// Chart is the top-level type from chartjs.
type Chart struct {
	Type    chartType `json:"type"`
	Label   string    `json:"label,omitempty"`
	Data    Data      `json:"data,omitempty"`
	Options Options   `json:"options,omitempty"`
}

// AddDataset adds a dataset to the chart.
func (c *Chart) AddDataset(d Dataset) {
	c.Data.Datasets = append(c.Data.Datasets, d)
}

func (c *Chart) AddAxis(axis Axis) {
	if c.Options.Scales == nil {
		c.Options.Scales = map[string]Axis{}
	}
	c.Options.Scales[axis.ID] = axis
}

// AddXAxis adds an x-axis to the chart and returns the ID of the added axis.
func (c *Chart) AddXAxis(x Axis) (string, error) {
	if x.ID == "" {
		x.ID = "x"
	}
	if x.Position == Left || x.Position == Right {
		return "", fmt.Errorf("chart: added x-axis to left or right")
	}

	c.AddAxis(x)
	return x.ID, nil
}

// AddYAxis adds an y-axis to the chart and return the ID of the added axis.
func (c *Chart) AddYAxis(y Axis) (string, error) {
	if y.ID == "" {
		y.ID = "y"
	}
	if y.Position == Top || y.Position == Bottom {
		return "", fmt.Errorf("chart: added y-axis to top or bottom")
	}
	c.AddAxis(y)
	return y.ID, nil
}
