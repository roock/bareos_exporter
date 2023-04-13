-- Reproduction of https://github.com/vierbergenlars/bareos_exporter/issues/10

INSERT INTO FileSet (fileset, md5, createtime, filesettext)
    VALUES
        ('FileSetX', 'd41d8cd98f00b204e9800998ecf8427e', NOW() - interval 14 day, ''),
        ('FileSetX', 'd41d8cd98f00b204e9800998ecf8427f', NOW() - interval 7 day, '');

INSERT INTO Job (job, name, type, level, jobstatus, schedtime, starttime, poolid)
    VALUES
        ('cx-fsx.1', 'cx-fsx', 'B', 'F', 'T', NOW() - interval 13 day, NOW() - interval 13 day, (select poolid from Pool WHERE name = 'Pool1')),
        ('cx-fsx.2', 'cx-fsx', 'B', 'F', 'T', NOW() - interval 6 day, NOW() - interval 6 day, (select poolid from Pool WHERE name = 'Pool1'));

UPDATE Job j SET
    clientid = (SELECT c.clientid from Client c WHERE c.name = 'c1-fd'),
    filesetid = (SELECT f.filesetid from FileSet f WHERE f.fileset = 'FileSetX' AND md5 = 'd41d8cd98f00b204e9800998ecf8427e'),
    endtime = j.starttime + interval 1 hour
    WHERE j.name = 'cx-fsx';

UPDATE Job j SET
    filesetid = (SELECT f.filesetid from FileSet f WHERE f.fileset = 'FileSetX' AND md5 = 'd41d8cd98f00b204e9800998ecf8427f')
    WHERE j.job = 'cx-fsx.2';
