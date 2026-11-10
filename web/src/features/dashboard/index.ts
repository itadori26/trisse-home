/** Public API of the dashboard feature. */
export { DashboardApi } from './adapter/api/dashboardApi';
export { DashboardUseCases } from './usecase/dashboardUseCases';
export type { Widget, WidgetConfig, WidgetType } from './domain/dashboard';
export { default as DashboardView } from './ui/DashboardView.svelte';
