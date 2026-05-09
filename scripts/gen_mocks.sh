#!/usr/bin/env bash
# Regenerate mocks for all *Repository and *Usecase interfaces.
set -euo pipefail

mockgen -source=internal/reservation/repository/type.go \
        -destination=mock/repository/reservation_repository_mock.go \
        -package=repository

mockgen -source=internal/reservation/usecase/type.go \
        -destination=mock/usecase/reservation_usecase_mock.go \
        -package=usecase

mockgen -source=internal/billing/repository/type.go \
        -destination=mock/repository/invoice_repository_mock.go \
        -package=repository

mockgen -source=internal/billing/usecase/type.go \
        -destination=mock/usecase/billing_usecase_mock.go \
        -package=usecase

echo "✓ mocks regenerated under mock/"
