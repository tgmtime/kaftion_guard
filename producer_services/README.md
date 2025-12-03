# tgmenver

commands
queries

her 3 saniyede bir kafkaya random veri gönderilecek (web servisin commands gibi hareket edicek dışarıdan gelen istekleri kafkaya iletiyor.)

yazılacak mesaj 
```json
{
  "eventId": "UUID",
  "eventType": "string",
  "timestamp": "ISO8601 Datetime",
  "data": {
    "key1": "value1",
    "key2": "value2"
  }
}
```