import requests
import json

def get_input(prompt, default=None, type_func=str):
    val = input(f"{prompt} [{default}]: " if default is not None else f"{prompt}: ")
    if not val and default is not None:
        return default
    try:
        return type_func(val)
    except ValueError:
        print("Invalid input, try again.")
        return get_input(prompt, default, type_func)

def main():
    print("--- RiskGo Portfolio Analyzer ---")
    
    ticker = get_input("Ticker (e.g. AAPL, TSLA)", "AAPL")
    quantity = get_input("Quantity", 100, float)
    beta = get_input("Beta (relative to SPY)", 1.0, float)
    volatility = get_input("Volatility (e.g. 0.3 for 30%)", 0.3, float)
    
    legs = []
    while True:
        print("\nAdd an Option Leg:")
        option_type = get_input("  Type (CALL/PUT)", "CALL").upper()
        strike = get_input("  Strike Price", 180.0, float)
        expiry = get_input("  Expiry (YYYY-MM-DD)", "2026-06-19")
        
        legs.append({
            "type": option_type,
            "strike": strike,
            "expiry": expiry
        })
        
        if get_input("Add another leg? (y/n)", "n").lower() != 'y':
            break

    payload = {
        "positions": [
            {
                "ticker": ticker,
                "quantity": quantity,
                "beta": beta,
                "legs": legs
            }
        ],
        "scenario_range": [-0.10, -0.05, 0, 0.05, 0.10],
        "volatility": volatility
    }

    print("\nSending request to RiskGo Gateway...")
    try:
        response = requests.post("http://localhost:3000/api/analyze_portfolio", json=payload)
        response.raise_for_status()
        results = response.json()

        print("\n" + "="*60)
        print(f"RISK REPORT: {ticker}")
        print("="*60)
        print(f"{'Scenario':<12} | {'PnL':<12} | {'Delta':<10} | {'Gamma':<10} | {'Theta':<10}")
        print("-" * 60)
        
        # Sort scenarios numerically for display
        ticker_results = results.get(ticker, {})
        sorted_scenarios = sorted(ticker_results.keys(), key=lambda x: float(x))
        
        for scene in sorted_scenarios:
            m = ticker_results[scene]
            print(f"{scene:<12} | {m['pnl']:>12.2f} | {m['delta']:>10.2f} | {m['gamma']:>10.4f} | {m['theta']:>10.2f}")
        
        print("="*60)

    except Exception as e:
        print(f"Error connecting to service: {e}")

if __name__ == "__main__":
    main()
