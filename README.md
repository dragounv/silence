# Sklízení velkého počtu semínek

## CLI interface

`silence [flags] command [command specific flags]`

### Commands

`run` - Spustí celý proces, defaultní výstup logů je stdout, na stderr se mohou objevit chybové hlášky

## Základní algoritmus

Vstupy:

- temlate Heritrix konfigurace
- semínka - `seeds.txt`
- konfigurace sklizní - `job.yaml | job.json`

Výstupy:

- warcy - `harvest-directory/*.warc`
- logy sklizně - `harvest-directory/logs/crawl/*.tar.gz`
- logy programu - `$WD/logs/process.log`
- zámek - soubor v tmp adresáři pro indikaci že proces již běží
- std.error - chybové výstupy pro uživatele

1. Inicializace procesu

    - zpracování příkazové řádky
    - pokud je vyvolaný příkaz pro zpracování sklizní
        - inicializace slog.Logger
        - kontrola zda již proces neběží (zda již neexistuje zámek) **TODO**
            - pokud ano ukoči process
        - vytvoření zámku **TODO**
        - inicializce App struktury

2. Nahrát templaty, semínka a konfiguraci sklizně

    - načtení konfigurace
        - parsování konfigurace
        - inicializace Job struktury
            - doplnění chybějících a defaultních hodnot
    - inicializace klienta pro komunikaci s heritrixem
        - ping Heritrixu
            - pokud ano, ukonči proces (process si spustí crawler samostatně a samostatně ho ukončí)
    - načtení templatů
    - načtení semínek

3. Rozdělit semínka a inicializovat sklizeň pro každý díl

    - rozdělit semínka na x dílů podle konfigurace
    - inicializovat Harvest struktury
        - vytvořit pracovní adresář
        - inicializovat FS
        - vytvořit crawler-beans.cxml
        - vytvořit seeds.txt
    - inicializace Heritrixu
        - znovu ping na heritrix
            - pokud odpoví ukonči proces (něco zapnulo heritrix?)
        - načti konfiguraci a přehraj script pro spuštění heritrixu
        - kontrolní ping na heritrix
            - zkoušej po dobu cca 90 sekund
                - pokud neodpoví, ukonči proces

4. Pro každou sklizeň ve frontě:

    - dequeue sklizeň z fronty
        - serializuj zbytek fronty pro případné obnovení
    - zkontroluj přítomnost rozpracované sklizně v adrsáři sklizně (ukazuje latest na existující soubor?)
        - pokud existuje, ukonči process
    - nahraj konfiguraci do adresáře sklizně
    - přehraj cyklus sklizně
        - build
        - start
        - unpause
        - pravidelné kontroly průběhu sklizně
            - pokud neodpoví zkus okamžitě znovu
                - pokud neodpoví třikrát
                    - pošli příkaz k ukončení sklizně
                    - pokus se o ukončení Heritrixu
                    - ukonči proces
            - pokud doba sklizně pekročí limit v konfiguraci
                - ukonči sklizeň
        - stop
        - terminate
    - úklid (proces by se měl pokusit o všechny tyto kroky i když dojde k chybě)
        - odeber koncovky .open z warců
        - spusť skript pro archivaci logů
        - odstraň
            - seeds.txt
            - crawler-beans.cxml
        - zkontroluj že soubory byly odstraněny a že logy se již nenacházejí v adresáři
            - pokud ne, ukonči proces
        - pokud doteď nastaly jakékoli jiné chyby, ukonči proces

5. Ukončit celý proces

    - ukonči heritrix
    - zavři všechny soubory, clienty a deinicializuj vše potřebné
    - překopíruj aktuální konfigurace, semínka a logy do archivu
    - os.Exit(0)
