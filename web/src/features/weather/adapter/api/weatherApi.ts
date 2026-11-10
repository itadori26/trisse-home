import type { HttpClient } from '../../../../shared/api-client/http';
import type { ForecastDTO, PlaceDTO } from '../../../../shared/api-client/generated/weather';
import type { Forecast, Place } from '../../domain/weather';
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
}
