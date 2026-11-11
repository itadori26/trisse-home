/** Builds the SVG polyline points for values scaled into a width × height box. */
export function sparklinePoints(values: number[], width: number, height: number, max?: number): string {
  if (values.length === 0) return '';
  const top = max ?? Math.max(...values, 1);
  const step = values.length > 1 ? width / (values.length - 1) : 0;
  return values
    .map((v, i) => {
      const x = values.length > 1 ? i * step : width / 2;
      const y = height - (Math.min(Math.max(v, 0), top) / top) * height;
      return `${round(x)},${round(y)}`;
    })
    .join(' ');
}

const round = (n: number) => Math.round(n * 10) / 10;
