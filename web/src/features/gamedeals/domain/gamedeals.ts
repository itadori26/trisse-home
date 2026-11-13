/** Settings of the game deals widget (#62). */
export interface GameDealsSettings {
  /** 64-bit Steam ID of a profile with a public wishlist; empty disables Steam. */
  steamId: string;
  country: string;
  minDiscount: number;
  epic: boolean;
}

/** A discounted wishlist game. */
export interface Deal {
  name: string;
  url: string;
  discountPct: number;
  finalPrice: string;
  endsAt: Date | null;
}

/** A free game on the Epic Games Store. */
export interface FreeGame {
  title: string;
  url: string;
  start: Date;
  end: Date;
}

/** What the widget shows. */
export interface GameDealsOverview {
  deals: Deal[];
  freeNow: FreeGame[];
  freeNext: FreeGame[];
  errors: string[];
}

/** Extracts a 64-bit Steam ID from a profile URL or returns the input. */
export function steamIdFrom(input: string): string {
  const m = /steamcommunity\.com\/profiles\/(\d{17})/.exec(input);
  return m?.[1] ?? input.trim();
}

/** Whether neither source is enabled. */
export function isUnconfigured(s: Pick<GameDealsSettings, 'steamId' | 'epic'>): boolean {
  return s.steamId === '' && !s.epic;
}
