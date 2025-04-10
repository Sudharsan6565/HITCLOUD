import boto3

dynamodb = boto3.resource('dynamodb', region_name='us-east-1')
table = dynamodb.Table('HybridHealAI_Tasks')

response = table.scan(Limit=5)

for item in response.get("Items", []):
    print(item)
