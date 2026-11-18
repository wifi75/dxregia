"""Elenca i testi italiani del pannello sysop che non hanno ancora una traduzione in src/adminTranslations.ts.

Uso, dalla cartella web/:  python scripts/missing_admin_strings.py
"""
import re
import sys

SRC = 'src/'
FILES = ['Auth', 'Analytics', 'ContactRequests', 'MailSettings', 'MojoLog', 'Partners', 'PasswordSettings',
         'BackupVerify', 'TelnetConsole', 'TelnetOptions', 'WebSettings', 'Setup', 'NodeMessage', 'NodeStartup',
         'NodeLists', 'NodeHops', 'NodeAccesses', 'Dashboard', 'BarList', 'ColumnChart']

table = open(SRC + 'adminTranslations.ts', encoding='utf-8').read()
keys = []
for m in re.finditer(r"^  \[('(?:[^'\\]|\\.)*'|\"(?:[^\"\\]|\\.)*\")", table, flags=re.M):
    key = m.group(1)[1:-1].replace("\\'", "'").replace('\\n', '\n')
    keys.append(re.sub(r'\s+', ' ', key).strip())
exact = {k for k in keys if '{}' not in k}
patterns = [re.compile('^' + '(.+?)'.join(re.escape(p) for p in k.split('{}')) + '$') for k in keys if '{}' in k]

code_like = re.compile(r'[=;(){}$]|\bconst\b|\blet\b|=>')
for name in FILES:
    text = open(SRC + name + '.svelte', encoding='utf-8').read()
    body = re.sub(r'<style.*?</style>', '', text, flags=re.S)
    found = set()
    for m in re.finditer(r'(?:title|placeholder|aria-label|alt|text)="([^"{}]*[A-Za-zÀ-ÿ][^"{}]*)"', body):
        found.add(m.group(1))
    for m in re.finditer(r'>([^<>]*[A-Za-zÀ-ÿ][^<>]*)<', body):
        found.add(m.group(1))
    for m in re.finditer(r"'([^'\n]{3,200})'", body):
        value = m.group(1)
        if ' ' in value and re.search(r'[a-zà-ÿ]{3}', value) and not value.startswith(('/', '.', '#')) and 'api/' not in value:
            found.add(value)
    missing = []
    for item in found:
        norm = re.sub(r'\s+', ' ', item).strip()
        if not norm or norm in exact or any(p.match(norm) for p in patterns):
            continue
        if code_like.search(norm) and not re.search(r'[à-ù]', norm):
            continue
        if re.fullmatch(r'[A-Za-z0-9_./:\-]+', norm):
            continue
        missing.append(norm)
    if missing:
        print('##', name)
        for norm in sorted(missing):
            print('  ', norm[:160])
sys.exit(0)
