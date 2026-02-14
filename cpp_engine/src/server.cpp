#include <iostream>
#include <memory>
#include <string>
#include <vector>
#include <cstdio>

#include <grpcpp/grpcpp.h>
#include "risk_engine.grpc.pb.h"

#include <ql/quantlib.hpp>
#include <boost/shared_ptr.hpp>
#include <boost/make_shared.hpp>

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;
using riskengine::RiskEngine;
using riskengine::ScenarioRequest;
using riskengine::ScenarioResponse;
using riskengine::OptionLeg;

namespace ql = QuantLib;

class RiskEngineServiceImpl final : public RiskEngine::Service {
    Status CalculateBetaScenario(ServerContext* context, const ScenarioRequest* request,
                                ScenarioResponse* response) override {
        try {
            double spot = request->spot_price();
            double rfr = request->risk_free_rate();
            double vol = request->volatility();
            double shock = request->scenario_pct_change();
            double beta = request->beta();

            double adjusted_spot = spot * (1.0 + (shock * beta));

            ql::Calendar calendar = ql::TARGET();
            ql::Date today = ql::Date::todaysDate();
            ql::Settings::instance().evaluationDate() = today;
            ql::DayCounter dayCounter = ql::Actual365Fixed();

            double total_pnl = 0.0;
            double total_delta = 0.0;
            double total_gamma = 0.0;
            double total_theta = 0.0;

            auto get_engine = [&](double current_spot) {
                ql::Handle<ql::Quote> underlyingH(boost::make_shared<ql::SimpleQuote>(current_spot));
                ql::Handle<ql::YieldTermStructure> flatTermStructure(
                    boost::make_shared<ql::FlatForward>(today, rfr, dayCounter));
                ql::Handle<ql::BlackVolTermStructure> flatVolTS(
                    boost::make_shared<ql::BlackConstantVol>(today, calendar, vol, dayCounter));
                
                auto bsmProcess = boost::make_shared<ql::BlackScholesMertonProcess>(
                    underlyingH, flatTermStructure, flatTermStructure, flatVolTS);
                return boost::make_shared<ql::AnalyticEuropeanEngine>(bsmProcess);
            };

            auto base_engine = get_engine(spot);
            auto shocked_engine = get_engine(adjusted_spot);

            for (const auto& leg : request->legs()) {
                ql::Option::Type type = (leg.type() == OptionLeg::CALL) ? ql::Option::Call : ql::Option::Put;
                
                int y, m, d;
                if (sscanf(leg.expiry().c_str(), "%d-%d-%d", &y, &m, &d) != 3) continue;
                ql::Date expiry(d, static_cast<ql::Month>(m), y);
                if (expiry <= today) continue;

                auto exercise = boost::make_shared<ql::EuropeanExercise>(expiry);
                auto payoff = boost::make_shared<ql::PlainVanillaPayoff>(type, leg.strike());

                ql::VanillaOption base_option(payoff, exercise);
                base_option.setPricingEngine(base_engine);
                double base_npv = base_option.NPV();

                ql::VanillaOption shocked_option(payoff, exercise);
                shocked_option.setPricingEngine(shocked_engine);
                double shocked_npv = shocked_option.NPV();

                total_pnl += (shocked_npv - base_npv) * leg.quantity();
                
                total_delta += base_option.delta() * leg.quantity();
                total_gamma += base_option.gamma() * leg.quantity();
                try {
                    total_theta += base_option.theta() * leg.quantity();
                } catch (...) {}
            }

            response->set_scenario_pnl(total_pnl);
            response->set_scenario_delta(total_delta);
            response->set_scenario_gamma(total_gamma);
            response->set_scenario_theta(total_theta);

            return Status::OK;
        } catch (const std::exception& e) {
            return Status(grpc::StatusCode::INTERNAL, e.what());
        }
    }
};

void RunServer() {
    std::string server_address("0.0.0.0:50051");
    RiskEngineServiceImpl service;

    ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);
    std::unique_ptr<Server> server(builder.BuildAndStart());
    if (server) {
        std::cout << "Server listening on " << server_address << std::endl;
        server->Wait();
    } else {
        std::cerr << "Failed to start server" << std::endl;
    }
}

int main(int argc, char** argv) {
    RunServer();
    return 0;
}
