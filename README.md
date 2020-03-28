# ugr ![](https://img.shields.io/github/go-mod/go-version/xnyo/ugr) ![](https://img.shields.io/github/license/xnyo/ugr)

## Informazioni
ugr è un bot Telegram per la gestione del servizio Spesa Sicura della città di Pisa: un
servizio per la consegna dei beni di prima necessità per anziani e bisognosi, nato
durante l'emergenza Coronavirus.

## Come funziona
Ci sono due tipologie di utenti: amministratori e volontari.
- Gli amministratori possono inserire gli ordini. A ogni ordine è assegnato:
  - Nominativo del destinatario
  - Indirizzo del destinatario
  - Zona della città
  - Numero di telefono
  - Scadenza (opzionale)
  - Note aggiuntive (optionale)
  - Foto aggiuntive, fino a un massimo di 10 (opzionale)
- I volontari scelgono la zona in cui si strovano e ricevono la lista degli ordini nella zona.
Possono assegnarsi un ordine e portarlo a termine.
- Gli amministratori possono invitare altri utenti come amministratori o volontari
- Le azioni rilevanti vengono tracciate su un canale dedicato

## Funzionalità
- Pannello amministratore
  - Gestione zone
  - Aggiunta ordini
  - Aggiunta nuovi volontari e amministratori
- Pannello volontario
- Supporto per MySQL (MariaDB) e SQLite
- Supporto per Docker & Docker Compose
- Logging su canale dedicato

## Impostazioni BotFather
- L'inline deve essere abilitato (`/setinline`, il testo del placeholder non è rilevante)
- Il bot non deve poter entrare nei gruppi (`/setjoingroups`, `Disable`)
- Il bot deve essere inserito nel canale di logging

## Licenza
MIT
