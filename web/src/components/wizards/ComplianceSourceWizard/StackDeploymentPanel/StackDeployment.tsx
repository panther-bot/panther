/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import { Text, Box, Spinner, Link, FormError, Flex } from 'pouncejs';
import React from 'react';
import { extractErrorMessage } from 'Helpers/utils';
import { useFormikContext } from 'formik';
import { pantherConfig } from 'Source/config';
import { WizardPanelWrapper } from 'Components/Wizard';
import { useGetComplianceCfnTemplate } from './graphql/getComplianceCfnTemplate.generated';
import { ComplianceSourceWizardValues } from '../ComplianceSourceWizard';

const StackDeployment: React.FC = () => {
  const { initialValues, values, setStatus } = useFormikContext<ComplianceSourceWizardValues>();
  const { data, loading, error } = useGetComplianceCfnTemplate({
    variables: {
      input: {
        awsAccountId: pantherConfig.AWS_ACCOUNT_ID,
        integrationLabel: values.integrationLabel,
        remediationEnabled: values.remediationEnabled,
        cweEnabled: values.cweEnabled,
      },
    },
  });

  const downloadRef = React.useCallback(
    node => {
      if (data && node) {
        const blob = new Blob([data.getComplianceIntegrationTemplate.body], {
          type: 'text/yaml;charset=utf-8',
        });

        const downloadUrl = URL.createObjectURL(blob);
        node.setAttribute('href', downloadUrl);
      }
    },
    [data]
  );

  const renderContent = () => {
    if (loading) {
      return (
        <Flex width={1} justify="center" my={5}>
          <Spinner size="medium" />
        </Flex>
      );
    }

    if (error) {
      return (
        <FormError>
          Couldn{"'"}t generate a Cloudformation template. {extractErrorMessage(error)}
        </FormError>
      );
    }

    const { stackName } = data.getComplianceIntegrationTemplate;
    const downloadTemplateLink = (
      <Link
        href="#"
        title="Download cloudformation template"
        download={`${stackName}.yml`}
        ref={downloadRef}
        onClick={() => setStatus({ cfnTemplateDownloaded: true })}
      >
        Download template
      </Link>
    );

    if (!initialValues.integrationId) {
      const cfnConsoleLink =
        `https://${pantherConfig.AWS_REGION}.console.aws.amazon.com/cloudformation/home?region=${pantherConfig.AWS_REGION}#/stacks/create/review` +
        `?templateURL=https://s3-us-west-2.amazonaws.com/panther-public-cloudformation-templates/panther-cloudsec-iam/v1.0.0/template.yml` +
        `&stackName=${stackName}` +
        `&param_MasterAccountRegion=${pantherConfig.AWS_REGION}` +
        `&param_MasterAccountId=${pantherConfig.AWS_ACCOUNT_ID}` +
        `&param_DeployCloudWatchEventSetup=${values.cweEnabled}` +
        `&param_DeployRemediation=${values.remediationEnabled}`;

      return (
        <Box fontSize="medium">
          <Text color="gray-300" mt={2} mb={2}>
            The quickest way to do it, is through the AWS console
          </Text>
          <Link
            external
            title="Launch cloudformation console"
            href={cfnConsoleLink}
            onClick={() => setStatus({ cfnTemplateDownloaded: true })}
          >
            Launch stack
          </Link>
          <Text color="gray-300" mt={10} mb={2}>
            Alternatively, you can download it and deploy it through the AWS CLI with the stack name{' '}
            <b>{stackName}</b>
          </Text>
          {downloadTemplateLink}
        </Box>
      );
    }

    return (
      <React.Fragment>
        <Box as="ol" fontSize="medium">
          <Box as="li" color="gray-300" mb={3}>
            1. {downloadTemplateLink}
          </Box>
          <Box as="li" color="gray-300" mb={3}>
            2. Log into your
            <Link
              external
              ml={1}
              title="Launch Cloudformation console"
              href={`https://${pantherConfig.AWS_REGION}.console.aws.amazon.com/cloudformation/home`}
            >
              Cloudformation console
            </Link>{' '}
            of the account <b>{values.awsAccountId}</b>
          </Box>
          <Box as="li" color="gray-300" mb={3}>
            3. Find the stack <b>{stackName}</b>
          </Box>
          <Box as="li" color="gray-300" mb={3}>
            4. Press <b>Update</b>, choose <b>Replace current template</b>
          </Box>
          <Box as="li" color="gray-300" mb={3}>
            5. Press <b>Next</b> and finally click on <b>Update</b>
          </Box>
        </Box>
        <Text color="gray-300" fontSize="medium" mt={10} mb={2}>
          Alternatively, you can update your stack through the AWS CLI
        </Text>
      </React.Fragment>
    );
  };

  return (
    <Box>
      <WizardPanelWrapper.Heading
        title="Deploy your configured stack"
        subtitle={`To proceed, you must deploy the generated Cloudformation template to the AWS account 
        ${values.awsAccountId}. 
        ${
          !initialValues.integrationId
            ? 'This will generate the necessary IAM Roles.'
            : 'This will update any previous IAM Roles.'
        }`}
      />
      {renderContent()}
    </Box>
  );
};

export default StackDeployment;
