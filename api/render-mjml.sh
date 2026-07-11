#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")"
mkdir -p internal/mailing/templates/rendered
npx --no-install mjml --config.allowIncludes true internal/mailing/templates/mjml/new-event.mjml -o internal/mailing/templates/rendered/new-event.html
npx --no-install mjml --config.allowIncludes true internal/mailing/templates/mjml/cancellation.mjml -o internal/mailing/templates/rendered/cancellation.html
