-- Reproduction of https://github.com/roock/bareos_exporter/issues/9

INSERT INTO Media (volumename, mediatype, firstwritten, lastwritten, labeldate, volstatus, poolid)
    VALUES
        ('Pool1-0100', 'NULL', NOW() - interval 1 day + interval 30 second, NULL, NOW() - interval 1 day, 'Full', (SELECT poolid from Pool WHERE name = 'Pool1'));
