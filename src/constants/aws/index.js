import Accounts from './accountsTypes';
import S3 from './s3Types';
import Costs from './costsTypes';
import Reports from './reportsTypes';
import Map from './mapTypes';

export default {
	...Accounts,
	...S3,
	...Costs,
	...Reports,
	...Map
};
