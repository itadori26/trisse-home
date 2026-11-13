/** Public API of the dashboard feature. */
export { DashboardApi } from './adapter/api/dashboardApi';
export { DashboardUseCases } from './usecase/dashboardUseCases';
export type { Tab, Widget, WidgetConfig, WidgetType } from './domain/dashboard';
export { firstWidgetConfig } from './domain/header';
export { default as DashboardView } from './ui/DashboardView.svelte';
