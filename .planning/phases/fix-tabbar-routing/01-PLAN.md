---
phase: fix
plan_id: 03
wave: 1
depends_on: [fix-01, fix-02]
files_modified:
  - frontend/src/App.tsx
  - frontend/src/pages/Login.tsx
autonomous: false
requirements:
  - "BUG-005: Após login, não há TabBar para navegar entre Registrar / Histórico / Perfil"
  - "BUG-006: Rota /history não existe no App.tsx"
  - "BUG-007: Login redireciona para /register que renderiza página sem TabBar"
---

# Plan 03: Fix TabBar routing — layout aninhado pós-login

## Objective

Restaurar o layout de navegação com TabBar (abas inferiores) após o login. Atualmente as rotas são planas — cada página renderiza isoladamente sem o TabBar.

## Tasks

### Task 1: Reestruturar rotas no App.tsx com layout aninhado

**type**: execute
**wave**: 1
**files_modified**:
  - frontend/src/App.tsx
  - frontend/src/pages/Login.tsx

<read_first>
- frontend/src/App.tsx
- frontend/src/components/TabBar.tsx
</read_first>

<acceptance_criteria>
- Rota `/login` renderiza `<Login />` sem TabBar
- Rotas `/register`, `/history`, `/profile` são aninhadas dentro de uma route layout que renderiza `<AuthGuard><TabBar /></AuthGuard>`
- `TabBar` usa `<Outlet />` para renderizar as páginas filhas
- A navegação após login redireciona para `/register` (com TabBar visível)
- Três abas visíveis no TabBar: Registrar (Pencil), Histórico (Clock), Perfil (User)
- Aba ativa destaca com `text-indigo-600`
- TypeScript check passa (`npx tsc --noEmit`)
- Testes passam (`pnpm test`)
</acceptance_criteria>

<action>
1. Atualizar `frontend/src/App.tsx`:
   - Adicionar import `import { History } from './pages/History'`
   - Estruturar rotas com layout aninhado:

```tsx
<Routes>
  <Route path="/login" element={<Login />} />
  <Route element={<AuthGuard><TabBar /></AuthGuard>}>
    <Route path="/register" element={<Register />} />
    <Route path="/history" element={<History />} />
    <Route path="/profile" element={<Profile />} />
  </Route>
  <Route path="*" element={<Login />} />
</Routes>
```

2. Atualizar `frontend/src/pages/Login.tsx`:
   - Redirecionamento pós-login continua para `/register` (que agora terá TabBar)
   - Nenhuma mudança adicional necessária no Login
</action>

---

## Verification

1. `cd frontend && npx tsc --noEmit` passa
2. `cd frontend && pnpm test` passa
3. Após login, usuário vê `/register` com TabBar (Registrar, Histórico, Perfil)
4. Navegação entre abas funciona
5. Rota `/login` não tem TabBar
