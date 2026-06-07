---
phase: fix
plan_id: fix-combobox-datepicker
wave: 1
depends_on: []
files_modified:
  - frontend/src/components/SentimentoCombobox.tsx
  - frontend/src/components/RegistrationForm.tsx
  - frontend/src/components/DateTimePicker.tsx
autonomous: false
requirements:
  - "BUG-008: SentimentoCombobox não lista os 13 sentimentos quando PouchDB está vazio"
  - "BUG-009: Native datetime-local popover não respeita o esquema de cores do tema"
---

# Plan: Fix SentimentoCombobox + Custom DateTimePicker

## Objective

Corrigir dois bugs visuais:

1. **SentimentoCombobox vazio** — Quando não há documentos `sentimento` no PouchDB, o combobox não exibe nenhuma opção. O usuário precisa digitar manualmente. Deve mostrar os 13 sentimentos padrão como fallback.

2. **Popover de data/hora nativo** — O `<input type="datetime-local">` usa o widget nativo do browser que ignora CSS custom properties. Substituir por um `DateTimePicker` customizado com `date-fns` estilizado com as cores do tema.

## Tasks

### Task 1: SentimentoCombobox — fallback para 13 sentimentos padrão

**type**: execute
**wave**: 1
**files_modified**:
  - frontend/src/components/SentimentoCombobox.tsx

<acceptance_criteria>
- Quando `getSentimentos()` retorna array vazio, o combobox exibe os 13 sentimentos padrão como opções selecionáveis
- Ao selecionar um sentimento do fallback, o `saveSentimento` é chamado para persistir no PouchDB
- O placeholder e filtro continuam funcionando normalmente
- Testes existentes continuam passando
</acceptance_criteria>

<action>
1. Adicionar constante `FALLBACK_EMOTIONS` com os 13 sentimentos (igual ao `SentimentoEditor.tsx`)
2. No estado inicial, se `sentimentos` estiver vazio após `getSentimentos()`, usar fallback para exibir as opções
3. No `handleSelect`, se o sentimento selecionado não existir no PouchDB, chamar `saveSentimento` para persistir
</action>

### Task 2: Criar DateTimePicker customizado com date-fns

**type**: execute
**wave**: 1
**files_modified**:
  - frontend/src/components/DateTimePicker.tsx
  - frontend/src/components/RegistrationForm.tsx

<acceptance_criteria>
- Substitui o `<input type="datetime-local">` por um componente estilizado com as cores do tema
- Exibe data e hora formatadas em pt-BR
- Ao clicar, abre um popover/overlay com calendário mensal navegável e campos de hora
- Usa `date-fns` para formatação e navegação entre meses
- Estilizado com `--primary`, `--secondary`, `--muted`, `--background`, etc.
- Valor selecionado é refletido no estado do formulário
- Testes existentes continuam passando
</acceptance_criteria>

<action>
1. Criar `frontend/src/components/DateTimePicker.tsx`:
   - Input que mostra data/hora formatada (DD/MM/AAAA HH:mm)
   - Overlay com:
     - Navegação de meses (< mês anterior | mês atual | próximo mês >)
     - Grid de dias destacando o dia selecionado com `--primary`
     - Inputs de hora e minuto
   - Fechar overlay ao selecionar dia + hora
   - `onChange` callback que retorna ISO string
2. Atualizar `RegistrationForm.tsx`:
   - Substituir `<input type="datetime-local">` por `<DateTimePicker>`
   - Manter estado `dataHora` como ISO string
</action>

---

## Verification

1. `cd frontend && npx tsc --noEmit` passa
2. `cd frontend && npm run test` passa — 94+ testes
3. SentimentoCombobox mostra 13 opções mesmo sem docs no PouchDB
4. DateTimePicker abre overlay estilizado com cores do tema
5. Selecionar data preenche o campo corretamente
6. Submit envia ISO string igual ao comportamento anterior
