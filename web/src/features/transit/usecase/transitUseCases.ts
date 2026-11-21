import type { Departure, Stop, TransitConfig } from '../domain/transit';

/** Output port: transit backend. */
export interface TransitGateway {
  stops(query: string): Promise<Stop[]>;
  stopsNear(lat: number, lon: number): Promise<Stop[]>;
  departures(stopId: string, count: number): Promise<Departure[]>;
}

/** Application logic of the transit feature. */
export class TransitUseCases {
  constructor(private readonly gateway: TransitGateway) {}

  /** Searches stops; short queries return nothing without a request. */
  searchStops(query: string): Promise<Stop[]> {
    const q = query.trim();
    return q.length < 2 ? Promise.resolve([]) : this.gateway.stops(q);
  }

  /** Stops nearest to a place, asked only when the person picks one (#694). */
  stopsNear(place: { lat: number; lon: number }): Promise<Stop[]> {
    return this.gateway.stopsNear(place.lat, place.lon);
  }

  departures(cfg: TransitConfig): Promise<Departure[]> {
    return this.gateway.departures(cfg.stopId, cfg.count);
  }
}
