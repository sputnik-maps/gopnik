namespace go gopnikrpc.baseservice
namespace cpp gopnikrpc.baseservice

service BaseService {
	bool status(),
	string version(),
	string config(),
	map<string,double> stat(),
}
