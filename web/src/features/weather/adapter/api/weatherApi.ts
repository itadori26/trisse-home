import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  ForecastDTO,
  HealthDTO,
  HealthOptionsDTO,
  PlaceDTO,
  WarningDTO,
} from '../../../../shared/api-client/generated/weather';
import type {
  Forecast,
  HealthOptions,
  HealthReport,
  PollenDay,
  Place,
  UVLevel,
  WeatherWarning,
} from '../../domain/weather';
import type { WeatherGateway } from '../../usecase/weatherUseCases';

/** WeatherGateway backed by the REST API. */
export class WeatherApi implements WeatherGateway {
  constructor(private readonly http: HttpClient) {}

  async forecast(lat: number, lon: number): Promise<Forecast> {
    const dto = await this.http.get<ForecastDTO>(`/api/weather?lat=${lat}&lon=${lon}`);
    return {
      temperature: dto.temperature,
      apparentTemperature: dto.apparentTemperature,
      humidity: dto.humidity,
      windSpeed: dto.windSpeed,
      windDirection: typeof dto.windDirection === 'number' ? dto.windDirection : null,
      code: dto.weatherCode,
      isDay: dto.isDay,
      updatedAt: new Date(dto.updatedAt),
      days: dto.days.map((d) => ({
        date: new Date(`${d.date}T12:00:00`),
        min: d.temperatureMin,
        max: d.temperatureMax,
        code: d.weatherCode,
        precipitationProbability: d.precipitationProbability,
      })),
    };
  }

  async searchPlaces(query: string): Promise<Place[]> {
    const dtos = await this.http.get<PlaceDTO[]>(`/api/weather/places?q=${encodeURIComponent(query)}`);
    return dtos.map((p) => ({ name: p.name, region: p.region, country: p.country, lat: p.lat, lon: p.lon }));
  }

  async warnings(lat: number, lon: number): Promise<WeatherWarning[]> {
    const dtos = await this.http.get<WarningDTO[]>(`/api/weather/warnings?lat=${lat}&lon=${lon}`);
    return dtos.map((w) => ({
      ...w,
      onset: w.onset ? new Date(w.onset) : null,
      expires: w.expires ? new Date(w.expires) : null,
    }));
  }

  async health(region: number, city: string, types: string[]): Promise<HealthReport> {
    const q = new URLSearchParams();
    if (region > 0) q.set('region', String(region));
    if (city) q.set('city', city);
    if (types.length) q.set('types', types.join(','));
    const d = await this.http.get<HealthDTO>(`/api/weather/health?${q}`);
    return {
      pollen: d.pollen
        ? { regionName: d.pollen.region.name, days: d.pollen.days.map((x) => ({ ...x }) as PollenDay) }
        : null,
      uv: d.uv
        ? {
            city: d.uv.city,
            today: d.uv.today,
            tomorrow: d.uv.tomorrow,
            todayLevel: d.uv.todayLevel as UVLevel,
            tomorrowLevel: d.uv.tomorrowLevel as UVLevel,
          }
        : null,
    };
  }

  async healthOptions(): Promise<HealthOptions> {
    const d = await this.http.get<HealthOptionsDTO>('/api/weather/health/options');
    return { regions: d.regions.map((r) => ({ ...r })), cities: [...d.cities] };
  }

  radarUrl(lat: number, lon: number, key: number): string {
    return `/api/weather/radar?${new URLSearchParams({ lat: String(lat), lon: String(lon), k: String(key) })}`;
  }
}
