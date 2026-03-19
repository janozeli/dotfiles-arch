Instale a statusline no projeto atual (diretório corrente).

1. Verifique se o kit de instalação existe:

```bash
ls ~/statusline-reinstall-script/gsd-statusline.js ~/statusline-reinstall-script/statusline
```

Se não existir, informe o usuário e interrompa.

2. Verifique se já está instalado:

```bash
ls .claude/hooks/gsd-statusline.js .claude/hooks/statusline 2>/dev/null
```

Se já existir, informe o usuário e pergunte se deseja sobrescrever.

3. Crie a pasta de hooks se não existir:

```bash
mkdir -p .claude/hooks
```

4. Copie os arquivos do kit de instalação para o diretório do projeto:

```bash
cp ~/statusline-reinstall-script/gsd-statusline.js .claude/hooks/gsd-statusline.js
cp -r ~/statusline-reinstall-script/statusline .claude/hooks/statusline
```

5. Confirme que a instalação foi bem-sucedida:

```bash
grep -q 'statusline/assembly' .claude/hooks/gsd-statusline.js && echo "OK: statusline instalada" || echo "ERRO: marker não encontrado"
```

Escopo: instala apenas no projeto atual (`.claude/hooks/` relativo ao diretório corrente), não afeta outros projetos.
