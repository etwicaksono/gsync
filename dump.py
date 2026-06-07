import urllib.request, json
url = 'https://rest.budgetbakers.com/wallet/openapi'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
resp = urllib.request.urlopen(req).read()
spec = json.loads(resp)

def get_schema(name):
    return spec['components']['schemas'][name]

def dump_schema(schema, indent=0):
    pad = ' ' * indent
    if '$ref' in schema:
        ref_name = schema['$ref'].split('/')[-1]
        return [pad + '$ref: ' + ref_name]
    out = []
    if 'type' in schema:
        out.append(pad + 'type: ' + str(schema.get('type')))
    if 'properties' in schema:
        out.append(pad + 'properties:')
        for k, v in schema['properties'].items():
            out.append(pad + '  ' + k + ':')
            out.extend(dump_schema(v, indent+4))
    if 'items' in schema:
        out.append(pad + 'items:')
        out.extend(dump_schema(schema['items'], indent+2))
    if 'allOf' in schema:
        out.append(pad + 'allOf:')
        for s in schema['allOf']:
            out.extend(dump_schema(s, indent+2))
    return out

for model in ['Account', 'Category', 'Record', 'PaginatedResponse', 'CreateRecordRequest', 'CreateRecordsResponse', 'BatchSummary', 'AmountWithCurrency']:
    print('=== ' + model + ' ===')
    try:
        print('\n'.join(dump_schema(get_schema(model))))
    except Exception as e:
        print('Error:', e)
