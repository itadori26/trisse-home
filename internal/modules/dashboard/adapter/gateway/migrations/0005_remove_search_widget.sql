-- The search widget is gone (the search field of the header does the same, #189, #221):
-- remove the widgets of this kind that were placed on tabs. Layouts saved as
-- templates keep it; creating a tab from a template skips kinds that are no
-- longer available.
DELETE FROM dashboard_widgets WHERE kind = 'search';
