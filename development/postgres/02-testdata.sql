INSERT INTO public.client (name, uname)
    VALUES
        ('c1-fd', 'Mock client'),
        ('c2-fd', 'Mock client');

INSERT INTO public.fileset (fileset, md5, createtime)
    VALUES
        ('FileSet1', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - interval '2 weeks'),
        ('FileSet2', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - interval '2 weeks'),
        ('FileSet3', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - interval '2 weeks');

INSERT INTO public.pool (name, pooltype, labelformat)
    VALUES
        ('Scratch', 'Scratch', '*'),
        ('Pool1', 'Backup', 'Pool1-'),
        ('Pool2', 'Backup', 'Pool1-');

UPDATE public.pool SET
    usecatalog=1,
    autoprune=1,
    recycle=1;

INSERT INTO public.media (volumename, mediatype, firstwritten, lastwritten, labeldate, volstatus)
    VALUES
        ('Pool1-0001', 'NULL', NOW() - interval '2 weeks' + interval '30s', NOW() - interval '2 weeks' + interval '1h', NOW() - interval '2 weeks', 'Full'),
        ('Pool1-0002', 'NULL', NOW() - interval '1 week 6 days' + interval '30s', NOW() - interval '1 week 6 days' + interval '1h', NOW() - interval '1 week 6 days', 'Used'),
        ('Pool2-0003', 'NULL', NOW() - interval '2 weeks' + interval '30s', NOW() - interval '2 weeks' + interval '1h', NOW() - interval '2 weeks', 'Full'),
        ('Pool2-0004', 'NULL', NOW() - interval '1 week 6 days' + interval '30s', NOW() - interval '1 week 6 days' + interval '1h', NOW() - interval '1 week 6 days', 'Used'),
        ('Pool1-0005', 'NULL', NOW() - interval '1 week 5 days' + interval '30s', NOW() - interval '1 week 5 days' + interval '1h', NOW() - interval '1 week 5 days', 'Used'),
        ('Pool2-0006', 'NULL', NOW() - interval '1 week 1 days' + interval '30s', NOW() - interval '1 week 1 days' + interval '1h', NOW() - interval '1 week 1 days', 'Used'),
        ('Pool1-0007', 'NULL', NOW() - interval '1 week' + interval '30s', NOW() - interval '1 week' + interval '1h', NOW() - interval '1 week', 'Used'),
        ('Pool2-0008', 'NULL', NOW() - interval '6 days' + interval '30s', NOW() - interval '6 days' + interval '1h', NOW() - interval '6 days', 'Used'),
        ('Pool1-0009', 'NULL', NOW() - interval '5 days' + interval '30s', NOW() - interval '5 days' + interval '1h', NOW() - interval '5 days', 'Used');

UPDATE public.media m SET poolid = (select poolid from public.pool p where p.name = substring(m.volumename from '(.*)-\d+'));

UPDATE public.media SET volretention = extract(epoch from interval '1 week');

INSERT INTO public.job (job, name, type, level, jobstatus, schedtime, starttime,  poolid)
    VALUES
        ('c1-fs1.1', 'c1-fs1', 'B', 'F', 'T', NOW() - interval '2 weeks', NOW() - interval '2 weeks', (select poolid from pool WHERE name = 'Pool1')),
        ('c1-fs2.1', 'c1-fs2', 'B', 'F', 'T', NOW() - interval '2 weeks', NOW() - interval '2 weeks', (select poolid from pool WHERE name = 'Pool1')),
        ('c2-fs1.2', 'c2-fs1', 'B', 'F', 'T', NOW() - interval '1 week 6 days', NOW() - interval '1 week 6 days', (select poolid from pool WHERE name = 'Pool1')),
        ('c2-fs1.3', 'c2-fs1', 'B', 'F', 'T', NOW() - interval '1 week 5 days', NOW() - interval '1 week 5 days', (select poolid from pool WHERE name = 'Pool2')),
        ('c1-fs1.4', 'c1-fs1', 'B', 'I', 'T', NOW() - interval '1 week 5 days', NOW() - interval '1 week 5 days', (select poolid from pool WHERE name = 'Pool2')),
        ('c2-fs1.5', 'c2-fs1', 'B', 'F', 'T', NOW() - interval '1 week 6 days', NOW() - interval '1 week 6 days', (select poolid from pool WHERE name = 'Pool1')),
        ('c1-fs1.6', 'c2-fs1', 'B', 'F', 'T', NOW() - interval '1 week 4 days', NOW() - interval '1 week 4 days', (select poolid from pool WHERE name = 'Pool2')),
        ('c1-fs1.7', 'c1-fs1', 'B', 'F', 'T', NOW() - interval '1 week 4 days', NOW() - interval '1 week 4 days', (select poolid from pool WHERE name = 'Pool1')),
        ('c1-fsx.8', 'c1-fsx', 'B', 'F', 'T', NOW() - interval '4 days', NOW() - interval '4 days', (select poolid from pool WHERE name = 'Pool1')),
        ('c1-fsx.9', 'c1-fsx', 'B', 'F', 'T', NOW() - interval '3 days', NOW() - interval '3 days', (select poolid from pool WHERE name = 'Pool1'));

UPDATE public.job j SET
    clientid = (SELECT c.clientid from public.client c WHERE c.name = CONCAT(substring(j.name from '([^-]+)-(.*)'), '-fd')),
    filesetid = (SELECT f.filesetid from public.fileset f WHERE f.fileset = CONCAT('FileSet', substring(j.name from ('[^-]+-fs(\d+)')))),
    endtime = j.starttime + interval '1h';

UPDATE public.job j SET
    filesetid = (SELECT f.filesetid from public.fileset f WHERE f.fileset = 'FileSet2')
    WHERE j.job = 'c1-fsx.8';
UPDATE public.job j SET
    filesetid = (SELECT f.filesetid from public.fileset f WHERE f.fileset = 'FileSet3')
    WHERE j.job = 'c1-fsx.9';

INSERT INTO public.jobmedia (jobid, mediaid)
    VALUES
        ((select jobid from public.job where job = 'c1-fs1.1'), (select mediaid from public.media where volumename = 'Pool1-0001')),
        ((select jobid from public.job where job = 'c1-fs2.1'), (select mediaid from public.media where volumename = 'Pool1-0001')),
        ((select jobid from public.job where job = 'c1-fs2.1'), (select mediaid from public.media where volumename = 'Pool1-0002')),
        ((select jobid from public.job where job = 'c2-fs1.2'), (select mediaid from public.media where volumename = 'Pool1-0002')),
        ((select jobid from public.job where job = 'c2-fs1.3'), (select mediaid from public.media where volumename = 'Pool2-0003')),
        ((select jobid from public.job where job = 'c1-fs1.4'), (select mediaid from public.media where volumename = 'Pool2-0003')),
        ((select jobid from public.job where job = 'c1-fs1.4'), (select mediaid from public.media where volumename = 'Pool2-0004')),
        ((select jobid from public.job where job = 'c2-fs1.5'), (select mediaid from public.media where volumename = 'Pool1-0005')),
        ((select jobid from public.job where job = 'c1-fs1.6'), (select mediaid from public.media where volumename = 'Pool2-0006')),
        ((select jobid from public.job where job = 'c1-fs1.6'), (select mediaid from public.media where volumename = 'Pool2-0008')),
        ((select jobid from public.job where job = 'c1-fsx.8'), (select mediaid from public.media where volumename = 'Pool1-0007')),
        ((select jobid from public.job where job = 'c1-fsx.9'), (select mediaid from public.media where volumename = 'Pool1-0007'));
