-- Reproduction of https://github.com/vierbergenlars/bareos_exporter/issues/9

INSERT INTO public.media (volumename, mediatype, firstwritten, lastwritten, labeldate, volstatus, poolid)
    VALUES
        ('Pool1-0100', 'NULL', NOW() - interval '1 day' + interval '30s', NULL, NOW() - interval '1 day', 'Full', (SELECT poolid from pool WHERE name = 'Pool1'));
