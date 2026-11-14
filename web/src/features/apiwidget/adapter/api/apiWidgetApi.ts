import type { HttpClient } from '../../../../shared/api-client/http';
import { serverMessage } from '../../../../shared/i18n';
import type { EndpointDTO, PreviewDTO, ValuesDTO } from '../../../../shared/api-client/generated/apiwidget';
import type { EndpointInputDTO } from '../../../../shared/api-client/generated/apiwidget-requests';
import type { Endpoint, EndpointInput, Preview, Values, ValueStatus } from '../../domain/apiwidget';
import type { ApiWidgetGateway } from '../../usecase/apiWidgetUseCases';

const toEndpoint = (d: EndpointDTO): Endpoint => ({
  ...d,
  extractions: d.extractions.map((x) => ({ ...x, warn: x.warn ?? null, crit: x.crit ?? null })),
});

const toValues = (d: ValuesDTO): Values => ({
  name: d.name,
  fetchedAt: new Date(d.fetchedAt),
  error: d.error ? (serverMessage(d.errorKey, d.errorParams ?? undefined) ?? d.error) : '',
  values: d.values.map((v) => ({ ...v, number: v.number ?? null, status: v.status as ValueStatus })),
});

const body = (i: EndpointInput): EndpointInputDTO => ({
  ...i,
  keepHeader: i.headerValue === '',
  extractions: i.extractions.map((x) => ({ ...x, warn: x.warn ?? undefined, crit: x.crit ?? undefined })),
});

/** ApiWidgetGateway backed by the REST API. */
export class ApiWidgetApi implements ApiWidgetGateway {
  constructor(private readonly http: HttpClient) {}

  async endpoints(): Promise<Endpoint[]> {
    return (await this.http.get<EndpointDTO[]>('/api/apiwidget/endpoints')).map(toEndpoint);
  }
  async create(input: EndpointInput): Promise<Endpoint> {
    return toEndpoint(await this.http.post<EndpointDTO>('/api/apiwidget/endpoints', body(input)));
  }
  async update(id: string, input: EndpointInput): Promise<Endpoint> {
    return toEndpoint(
      await this.http.put<EndpointDTO>(`/api/apiwidget/endpoints/${encodeURIComponent(id)}`, body(input)),
    );
  }
  remove(id: string): Promise<void> {
    return this.http.delete(`/api/apiwidget/endpoints/${encodeURIComponent(id)}`);
  }
  async preview(input: EndpointInput, id?: string): Promise<Preview> {
    const d = await this.http.post<PreviewDTO>('/api/apiwidget/preview', {
      id: id ?? '',
      endpoint: body(input),
    });
    return { ...toValues(d), body: d.body };
  }
  async values(id: string): Promise<Values> {
    return toValues(
      await this.http.get<ValuesDTO>(`/api/apiwidget/endpoints/${encodeURIComponent(id)}/values`),
    );
  }
}
