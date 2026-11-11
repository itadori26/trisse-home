import qrcode from 'qrcode-generator';

/** Renders text as a QR code and returns it as a data: URL image (GIF). */
export function qrDataUrl(text: string, cellSize = 5): string {
  const qr = qrcode(0, 'M');
  qr.addData(text);
  qr.make();
  return qr.createDataURL(cellSize, 2);
}
