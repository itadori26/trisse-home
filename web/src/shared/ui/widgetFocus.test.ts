import { describe, expect, it, vi } from 'vitest';
import { onWidgetFocusRequest, pickWidget, requestWidgetFocus } from './widgetFocus';

describe('widget focus (#52)', () => {
  const tabs = [
    { id: 'home', widgets: [{ id: 'w1', kind: 'clock' }] },
    {
      id: 'server',
      widgets: [
        { id: 'w2', kind: 'status-overview' },
        { id: 'w3', kind: 'services' },
      ],
    },
    { id: 'more', widgets: [{ id: 'w4', kind: 'services' }] },
  ];

  it('prefers the active tab, then the preferred kind', () => {
    expect(pickWidget(tabs, 'more', ['services', 'status-overview'])).toEqual({
      tabId: 'more',
      widgetId: 'w4',
    });
    expect(pickWidget(tabs, 'home', ['services', 'status-overview'])).toEqual({
      tabId: 'server',
      widgetId: 'w3',
    });
    expect(pickWidget(tabs, 'home', ['todos'])).toBeNull();
  });

  it('delivers requests to subscribers until they unsubscribe', () => {
    const handler = vi.fn();
    const off = onWidgetFocusRequest(handler);
    requestWidgetFocus(['todos']);
    off();
    requestWidgetFocus(['waste']);
    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith(['todos']);
  });
});
