import {
  type Endpoint,
  type EndpointInput,
  type Preview,
  validateInput,
  type Values,
} from '../domain/apiwidget';

/** Output port: JSON widget backend. */
export interface ApiWidgetGateway {
  endpoints(): Promise<Endpoint[]>;
  create(input: EndpointInput): Promise<Endpoint>;
  update(id: string, input: EndpointInput): Promise<Endpoint>;
  remove(id: string): Promise<void>;
  preview(input: EndpointInput, id?: string): Promise<Preview>;
  values(id: string): Promise<Values>;
}

/** Application logic of the JSON widget (#48). */
export class ApiWidgetUseCases {
  constructor(private readonly gateway: ApiWidgetGateway) {}

  endpoints(): Promise<Endpoint[]> {
    return this.gateway.endpoints();
  }

  /** Validates and saves; with stored set the endpoint is updated (an empty header value keeps the stored one). */
  save(input: EndpointInput, stored?: Endpoint): Promise<Endpoint> {
    const clean = this.clean(input);
    const problem = validateInput(clean, stored?.hasHeader ?? false);
    if (problem) return Promise.reject(new Error(problem));
    return stored ? this.gateway.update(stored.id, clean) : this.gateway.create(clean);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** Runs the request of the form without saving it. */
  preview(input: EndpointInput, stored?: Endpoint): Promise<Preview> {
    const clean = this.clean(input);
    const problem = validateInput(clean, stored?.hasHeader ?? false);
    if (problem) return Promise.reject(new Error(problem));
    return this.gateway.preview(clean, stored?.id);
  }

  values(id: string): Promise<Values> {
    return this.gateway.values(id);
  }

  private clean(i: EndpointInput): EndpointInput {
    return {
      ...i,
      name: i.name.trim(),
      url: i.url.trim(),
      headerName: i.headerName.trim(),
      extractions: i.extractions.map((x) => ({
        ...x,
        label: x.label.trim(),
        path: x.path.trim(),
        unit: x.unit.trim(),
        warn: Number.isFinite(x.warn) ? x.warn : null,
        crit: Number.isFinite(x.crit) ? x.crit : null,
      })),
    };
  }
}
