import os
import sys
import re
import subprocess

def process_issue_body(issue_body, pagerduty_score_threshold):
    affected_areas_pattern = r'###\s*Affected\s*areas\s*.*?\((.*?)\)'
    additional_affected_areas_pattern = r'###\s*Additional\s*affected\s*areas\s*.*?\((.*?)\)'
    prod_non_prod_pattern = r'###\s*Prod/Non-prod\s*environments\?\s*.*?\(x(\d+)\)'
    user_unblocked_pattern = r'###\s*Is\s*User\s*unblocked\?\s*.*?\(x(\d+)\)'
    user_unblocked_reason_pattern = r'###\s*How\s*was\s*the\s*user\s*un-blocked\?\s*.*?\(x([\d.]+)\)'

    affected_areas_match = re.search(affected_areas_pattern, issue_body, re.IGNORECASE)
    additional_affected_areas_match = re.search(additional_affected_areas_pattern, issue_body, re.IGNORECASE)
    prod_non_prod_match = re.search(prod_non_prod_pattern, issue_body, re.IGNORECASE)
    user_unblocked_match = re.search(user_unblocked_pattern, issue_body, re.IGNORECASE)
    user_unblocked_reason_match = re.search(user_unblocked_reason_pattern, issue_body, re.IGNORECASE)

    affected_areas = int(affected_areas_match.group(1).strip()) if affected_areas_match else 0
    additional_affected_areas = int(additional_affected_areas_match.group(1).strip()) if additional_affected_areas_match else 0
    prod_non_prod = int(prod_non_prod_match.group(1).strip()) if prod_non_prod_match else 0
    user_unblocked = int(user_unblocked_match.group(1).strip()) if user_unblocked_match else 0
    user_unblocked_reason = user_unblocked_reason_match.group(1).strip() if user_unblocked_reason_match else "0"

    try:
        user_unblocked_reason = int(user_unblocked_reason)
    except ValueError:
        try:
            user_unblocked_reason = float(user_unblocked_reason)
        except ValueError:
            user_unblocked_reason = 0

    print("Affected areas:", affected_areas)
    print("Additional affected areas:", additional_affected_areas)
    print("Prod/Non-prod environments?:", prod_non_prod)
    print("Is User unblocked?:", user_unblocked)
    print("How was the user un-blocked?:", user_unblocked_reason)
 
    if any(value == 0 for value in [affected_areas, prod_non_prod, user_unblocked]):
        print("One or more required values are missing. Exiting...")
        sys.exit(0)

    if user_unblocked_reason == 0:
        user_unblocked_reason = 1

    if user_unblocked_reason == 3:
        try:
            result = subprocess.run(['gh', 'issue', 'edit', str(issue_number), '--add-label', 'urgent'], capture_output=True, check=True, text=True)
            print("urgent label added to issue", issue_number)
        except subprocess.CalledProcessError as e:
            print(e.stderr)

    final_score = (affected_areas + additional_affected_areas) * prod_non_prod * user_unblocked * user_unblocked_reason
    print("Final Score:", final_score)

    comment = f"Final Score: {final_score}"
    try:
        result1 = subprocess.run(['gh', 'issue', 'comment', str(issue_number), '--body', comment], capture_output=True, check=True, text=True)
        print("Final score commented on issue", issue_number)
    except subprocess.CalledProcessError as e:
        print(e.stderr)

    return final_score

token = os.environ.get('MY_ACCESS_TOKEN')
subprocess.run(['gh', 'auth', 'login', '--with-token'], input=token, text=True, capture_output=True)

issue_body = os.environ.get('ISSUE_BODY')
issue_number=os.environ.get('ISSUE_NUMBER')

pagerduty_score_threshold = 300
final_score = process_issue_body(issue_body, pagerduty_score_threshold)

if final_score <= pagerduty_score_threshold:
    try:
        result = subprocess.run(['gh', 'issue', 'edit', str(issue_number), '--remove-label', 'pager-duty'])
        print("bug label removed from issue", issue_number)
    except subprocess.CalledProcessError as e:
        print(e)

