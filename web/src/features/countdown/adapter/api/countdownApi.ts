import type { HttpClient } from '../../../../shared/api-client/http';
import type { CountdownDTO } from '../../../../shared/api-client/generated/countdown';
import type { CountdownInputDTO } from '../../../../shared/api-client/generated/countdown-requests';
import type { Countdown, CountdownInput } from '../../domain/countdown';
import type { CountdownGateway } from '../../usecase/countdownUseCases';

const toCountdown = (d: CountdownDTO): Countdown => ({
  id: d.id,
  title: d.title,
  emoji: d.emoji,
  date: d.date,
  time: d.time,
});

const toBody = (i: CountdownInput): CountdownInputDTO => ({
  title: i.title,
  emoji: i.emoji,
  date: i.date,
  time: i.time,
});

/** CountdownGateway backed by the REST API. */
export class CountdownApi implements CountdownGateway {
  constructor(private readonly http: HttpClient) {}

  async list(graceDays: number): Promise<Countdown[]> {
    return (await this.http.get<CountdownDTO[]>(`/api/countdowns?graceDays=${graceDays}`)).map(toCountdown);
  }

  async create(input: CountdownInput): Promise<Countdown> {
    return toCountdown(await this.http.post<CountdownDTO>('/api/countdowns', toBody(input)));
  }

  async update(id: string, input: CountdownInput): Promise<Countdown> {
    return toCountdown(
      await this.http.put<CountdownDTO>(`/api/countdowns/${encodeURIComponent(id)}`, toBody(input)),
    );
  }

  remove(id: string): Promise<void> {
    return this.http.delete(`/api/countdowns/${encodeURIComponent(id)}`);
  }
}
