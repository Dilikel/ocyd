package hwmon

// CToF accepts a temperature value in Celsius (int) and returns its equivalent
// converted to Fahrenheit (int).
func CToF(celsius int) int {
	return (celsius * 9 / 5) + 32
}
