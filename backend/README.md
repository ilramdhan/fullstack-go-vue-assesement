Scaffold demonstrating entity / repository / usecase / service separation.

generate openapi:

```bash
make openapi-gen
```

generate JWT_SECRET:

```bash
make gen-secret
```

Run server:

```bash
cp env.sample .env
make tool-openapi
make openapi-gen
make dep
make gen-secret
make run
```

API:

- POST /dashboard/v1/auth/login {email,password}
- GET /dashboard/v1/payments?sort=sort,status=status,id=id
- PUT /dashboard/v1/payment/{id}/review
