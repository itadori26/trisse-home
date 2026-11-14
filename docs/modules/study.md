# Study progress

**What it does:** A manually maintained list of your study modules with ECTS, semester, status (open, in progress, passed) and grade (#57). The widget shows the passed credits as a progress bar against the credits of the whole degree, the ECTS-weighted grade average and the modules in progress.

## Configuration

Settings → Study progress: the total ECTS of the degree (default 180) and the modules. Grades use the German scale 1.0 (best) to 4.0 (passed) and are rounded to one decimal; passed modules without a grade (pass/fail, e.g. an internship) count towards the credits but not towards the average.

## Outbound requests

None.

## API

`GET/POST /api/study/modules`, `PUT/DELETE /api/study/modules/{id}`, `GET /api/study/progress`, `GET/PUT /api/study/settings` – see `api/modules/study.yaml`.
