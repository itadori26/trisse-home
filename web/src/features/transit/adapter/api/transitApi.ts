import type { HttpClient } from '../../../../shared/api-client/http';
import type { DepartureDTO, StopDTO } from '../../../../shared/api-client/generated/transit';
import type { Departure, Stop } from '../../domain/transit';
import type { TransitGateway } from '../../usecase/transitUseCases';

/** TransitGateway backed by the REST API. */
export class TransitApi implements TransitGateway {
  constructor(private readonly http: HttpClient) {}

  async stops(query: string): Promise<Stop[]> {
    return (await this.http.get<StopDTO[]>(`/api/transit/stops?${new URLSearchParams({ q: query })}`)).map(
      (s) => ({ ...s }),
    );
  }

  async departures(stopId: string, count: number): Promise<Departure[]> {
    const q = new URLSearchParams({ stop: stopId, n: String(count) });
    const dtos = await this.http.get<DepartureDTO[]>(`/api/transit/departures?${q}`);
    return dtos.map((d) => ({
      ...d,
      planned: new Date(d.planned),
      expected: d.expected ? new Date(d.expected) : null,
    }));
  }
}
