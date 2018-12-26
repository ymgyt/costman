=======
costman
=======

costman reports cloud vendor(aws, gcp) costs to slack.
it is intended to run as cronjob.


IAM Policy
==========

costman aws scenario require costexplorer permissions like that.

.. code-block:: json

   {
       "Version": "2012-10-17",
       "Statement": [
           {
               "Effect": "Allow",
               "Action": [
                   "ce:*"
               ],
               "Resource": [
                   "*"
               ]
           }
       ]
   }