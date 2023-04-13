-- Reproduction of https://github.com/vierbergenlars/bareos_exporter/issues/10

INSERT INTO public.fileset (fileset, md5, createtime)
    VALUES
        ('FileSetX', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - interval '2 weeks'),
        ('FileSetX', 'd41d8cd98f00b204e9800998ecf8427f', NOW() - interval '1 week');

INSERT INTO public.job (job, name, type, level, jobstatus, schedtime, starttime, poolid)
    VALUES
        ('cx-fsx.1', 'cx-fsx', 'B', 'F', 'T', NOW() - interval '1 week 6 days', NOW() - interval '1 week 6 days', (select poolid from pool WHERE name = 'Pool1')),
        ('cx-fsx.2', 'cx-fsx', 'B', 'F', 'T', NOW() - interval '6 days', NOW() - interval '6 days', (select poolid from pool WHERE name = 'Pool1'));

UPDATE public.job j SET
    clientid = (SELECT c.clientid from public.client c WHERE c.name = 'c1-fd'),
    filesetid = (SELECT f.filesetid from public.fileset f WHERE f.fileset = 'FileSetX' AND md5 = 'd41d8cd98f00b204e9800998ecf8427e'),
    endtime = j.starttime + interval '1h'
    WHERE j.name = 'cx-fsx';

UPDATE public.job j SET
    filesetid = (SELECT f.filesetid from public.fileset f WHERE f.fileset = 'FileSetX' AND md5 = 'd41d8cd98f00b204e9800998ecf8427f')
    WHERE j.job = 'cx-fsx.2';
