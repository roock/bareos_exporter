INSERT INTO Client (Name, Uname)
    VALUES
        ('c1-fd', 'Mock client'),
        ('c2-fd', 'Mock client');

INSERT INTO FileSet (FileSet, MD5, CreateTime, FileSetText)
    VALUES
        ('FileSet1', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - INTERVAL 2 WEEK, ''),
        ('FileSet2', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - INTERVAL 2 WEEK, ''),
        ('FileSet3', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - INTERVAL 2 WEEK, '');

INSERT INTO Pool (Name, PoolType, LabelFormat)
    VALUES
        ('Scratch', 'Scratch', '*'),
        ('Pool1', 'Backup', 'Pool1-'),
        ('Pool2', 'Backup', 'Pool1-');

UPDATE Pool SET
    UseCatalog=1,
    AutoPrune=1,
    Recycle=1;

INSERT INTO Media (volumename, mediatype, firstwritten, lastwritten, labeldate, volstatus)
    VALUES
        ('Pool1-0001', 'NULL', NOW() - interval 2 week + interval 30 second, NOW() - interval 2 week + interval 1 hour, NOW() - interval 2 week, 'Full'),
        ('Pool1-0002', 'NULL', NOW() - interval 13 day + interval 30 second, NOW() - interval 13 day + interval 1 hour, NOW() - interval 13 day, 'Used'),
        ('Pool2-0003', 'NULL', NOW() - interval 2 week + interval 30 second, NOW() - interval 2 week + interval 1 hour, NOW() - interval 2 week, 'Full'),
        ('Pool2-0004', 'NULL', NOW() - interval 13 day + interval 30 second, NOW() - interval 13 day + interval 1 hour, NOW() - interval 13 day, 'Used'),
        ('Pool1-0005', 'NULL', NOW() - interval 12 day + interval 30 second, NOW() - interval 12 day + interval 1 hour, NOW() - interval 12 day, 'Used'),
        ('Pool2-0006', 'NULL', NOW() - interval 8 day + interval 30 second, NOW() - interval 8 day + interval 1 hour, NOW() - interval 8 day, 'Used'),
        ('Pool1-0007', 'NULL', NOW() - interval 7 day + interval 30 second, NOW() - interval 7 day + interval 1 hour, NOW() - interval 7 day, 'Used'),
        ('Pool2-0008', 'NULL', NOW() - interval 6 day + interval 30 second, NOW() - interval 6 day + interval 1 hour, NOW() - interval 6 day, 'Used'),
        ('Pool1-0009', 'NULL', NOW() - interval 5 day + interval 30 second, NOW() - interval 5 day + interval 1 hour, NOW() - interval 5 day, 'Used');

UPDATE Media m SET poolid = (select poolid from Pool p where p.name = substring(m.volumename from 1 for 5));

UPDATE Media SET volretention = 604800; -- 1 week

INSERT INTO Job (job, name, type, level, jobstatus, schedtime, starttime,  poolid)
    VALUES
        ('c1-fs1.1', 'c1-fs1', 'B', 'F', 'T', NOW() - interval 2 week, NOW() - interval 2 week, (select poolid from Pool WHERE name = 'Pool1')),
        ('c1-fs2.1', 'c1-fs2', 'B', 'F', 'T', NOW() - interval 2 week, NOW() - interval 2 week, (select poolid from Pool WHERE name = 'Pool1')),
        ('c2-fs1.2', 'c2-fs1', 'B', 'F', 'T', NOW() - interval 13 day, NOW() - interval 13 day, (select poolid from Pool WHERE name = 'Pool1')),
        ('c2-fs1.3', 'c2-fs1', 'B', 'F', 'T', NOW() - interval 12 day, NOW() - interval 12 day, (select poolid from Pool WHERE name = 'Pool2')),
        ('c1-fs1.4', 'c1-fs1', 'B', 'I', 'T', NOW() - interval 12 day, NOW() - interval 12 day, (select poolid from Pool WHERE name = 'Pool2')),
        ('c2-fs1.5', 'c2-fs1', 'B', 'F', 'T', NOW() - interval 13 day, NOW() - interval 13 day, (select poolid from Pool WHERE name = 'Pool1')),
        ('c1-fs1.6', 'c2-fs1', 'B', 'F', 'T', NOW() - interval 11 day, NOW() - interval 11 day, (select poolid from Pool WHERE name = 'Pool2')),
        ('c1-fs1.7', 'c1-fs1', 'B', 'F', 'T', NOW() - interval 11 day, NOW() - interval 11 day, (select poolid from Pool WHERE name = 'Pool1')),
        ('c1-fsx.8', 'c1-fsx', 'B', 'F', 'T', NOW() - interval 4 day, NOW() - interval 4 day, (select poolid from Pool WHERE name = 'Pool1')),
        ('c1-fsx.9', 'c1-fsx', 'B', 'F', 'T', NOW() - interval 3 day, NOW() - interval 3 day, (select poolid from Pool WHERE name = 'Pool1'));

UPDATE Job j SET
    clientid = (SELECT c.clientid from Client c WHERE c.name = CONCAT(substring(j.name from 1 for 2), '-fd')),
    filesetid = (SELECT f.filesetid from FileSet f WHERE f.fileset = CONCAT('FileSet', substring(j.name from 6))),
    endtime = j.starttime + interval 1 hour;

UPDATE Job j SET
    filesetid = (SELECT f.filesetid from FileSet f WHERE f.fileset = 'FileSet2')
    WHERE j.job = 'c1-fsx.8';
UPDATE Job j SET
    filesetid = (SELECT f.filesetid from FileSet f WHERE f.fileset = 'FileSet3')
    WHERE j.job = 'c1-fsx.9';

INSERT INTO JobMedia (jobid, mediaid)
    VALUES
        ((select jobid from Job where job = 'c1-fs1.1'), (select mediaid from Media where volumename = 'Pool1-0001')),
        ((select jobid from Job where job = 'c1-fs2.1'), (select mediaid from Media where volumename = 'Pool1-0001')),
        ((select jobid from Job where job = 'c1-fs2.1'), (select mediaid from Media where volumename = 'Pool1-0002')),
        ((select jobid from Job where job = 'c2-fs1.2'), (select mediaid from Media where volumename = 'Pool1-0002')),
        ((select jobid from Job where job = 'c2-fs1.3'), (select mediaid from Media where volumename = 'Pool2-0003')),
        ((select jobid from Job where job = 'c1-fs1.4'), (select mediaid from Media where volumename = 'Pool2-0003')),
        ((select jobid from Job where job = 'c1-fs1.4'), (select mediaid from Media where volumename = 'Pool2-0004')),
        ((select jobid from Job where job = 'c2-fs1.5'), (select mediaid from Media where volumename = 'Pool1-0005')),
        ((select jobid from Job where job = 'c1-fs1.6'), (select mediaid from Media where volumename = 'Pool2-0006')),
        ((select jobid from Job where job = 'c1-fs1.6'), (select mediaid from Media where volumename = 'Pool2-0008')),
        ((select jobid from Job where job = 'c1-fsx.8'), (select mediaid from Media where volumename = 'Pool1-0007')),
        ((select jobid from Job where job = 'c1-fsx.9'), (select mediaid from Media where volumename = 'Pool1-0007'));
